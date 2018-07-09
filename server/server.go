// Package server contains a wrapper around the standard http.Server that
// encodes some useful behavior. Primarily, we set a number of timeouts to
// prevent issues as a result of badly behaved clients, and we listen for
// various interrupts and attempt to shutdown gracefully instead of terminating
// immediately.
package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/frazercomputing/f4/log"
)

const (
	// ReadTimeout is the maximum duration for reading the entire request,
	// including the body.
	ReadTimeout = 5 * time.Second

	// WriteTimeout is the maximum duration before timing out writes of the
	// response.
	WriteTimeout = 10 * time.Second

	// ShutdownTimeout is the maximum time we wait for in-flight requests to
	// finish before terminating the server.
	ShutdownTimeout = 5 * time.Second

	// MaxHeaderBytes controls the maximum number of bytes the server will read
	// parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body.
	MaxHeaderBytes = 1 << 20
)

// Server is a thin wrapper around the default http.Server.
type Server struct {
	addr     string
	server   http.Server
	shutdown time.Duration
	out, err io.Writer
}

// New returns a new Server with sane timeouts, and the supplied address and
// handler.
func New(addr string, h http.Handler, opts ...Option) *Server {
	s := &Server{
		addr: addr,
		server: http.Server{
			Addr:           addr,
			Handler:        h,
			ReadTimeout:    ReadTimeout,
			WriteTimeout:   WriteTimeout,
			MaxHeaderBytes: MaxHeaderBytes,
		},
		shutdown: ShutdownTimeout,
		out:      os.Stdout,
		err:      os.Stderr,
	}

	for _, opt := range opts {
		s = opt(s)
	}

	return s
}

// Option is passed to New to modify the default parameters for things like
// timeouts, output channels, etc.
type Option func(s *Server) *Server

// WithReadTimeout modifies the server to set the read timeout to the provided
// value.
func WithReadTimeout(to time.Duration) Option {
	return func(s *Server) *Server {
		s.server.ReadTimeout = to
		return s
	}
}

// WithWriteTimeout modifies the server to set the write timeout to the provided
// value.
func WithWriteTimeout(to time.Duration) Option {
	return func(s *Server) *Server {
		s.server.WriteTimeout = to
		return s
	}
}

// WithShutdown modifies the server to set the shutdown timeout to the provided
// value.
func WithShutdown(to time.Duration) Option {
	return func(s *Server) *Server {
		s.shutdown = to
		return s
	}
}

// WithMaxHeaderBytes modifies the server to set the maximum header bytes to the
// provided value.
func WithMaxHeaderBytes(n int) Option {
	return func(s *Server) *Server {
		s.server.MaxHeaderBytes = n
		return s
	}
}

// WithOutputWriter modifies the server to set the output writer to the provided
// value.
//
// The server uses this writer for non-error messages.
func WithOutputWriter(w io.Writer) Option {
	return func(s *Server) *Server {
		s.out = w
		return s
	}
}

// WithErrorWriter modifies the server to set the error writer to the provided
// value.
//
// The server uses this writer for any errors produced by the server.
func WithErrorWriter(w io.Writer) Option {
	return func(s *Server) *Server {
		s.err = w
		return s
	}
}

// ListenAndServe starts the wrapped server and listens for a number of
// interrupts which will trigger a shutdown. The shutdown attempts to be
// graceful and wait for in-flight requests to finish, but will shutdown
// forcefully if the timeout is exceeded.
func (s *Server) ListenAndServe(ctx context.Context) error {
	log.Trace(ctx, "f4/http/server/Server.ListenAndServe")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fmt.Fprintf(s.out, "listening on %s...\n", s.addr)
		fmt.Fprintf(s.err, s.server.ListenAndServe().Error())
	}()

	osSignals := make(chan os.Signal)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-osSignals

	ctx, cancel := context.WithTimeout(ctx, s.shutdown)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Fprintf(s.err, "shutdown timed out after %s: %v", s.shutdown, err)
		if err := s.server.Close(); err != nil {
			fmt.Fprintf(s.err, "error killing server: %v", err)
			return err
		}
	}

	wg.Wait()

	return nil
}
