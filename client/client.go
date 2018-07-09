// Package client provides a wrapper around Go's default http client that sets
// sane defaults for the timeout, since Go does not.
package client

import (
	"net/http"
	"time"
)

// DefaultTimeout is 5s and is used if no other timeout is provided.
const DefaultTimeout = 5 * time.Second

// Client is a wrapper around the default Go http client that sets a sane
// default for timeout.
type Client struct {
	*http.Client
}

// Option is passed to New to modify the default parameters for things like
// timeout, transport, etc.
type Option func(c *Client) *Client

// WithTimeout returns an Option that sets the client timeout to the provided
// value.
func WithTimeout(to time.Duration) Option {
	return func(c *Client) *Client {
		c.Timeout = to
		return c
	}
}

// WithTransport returns an Option that sets the client RoundTripper to the
// provided value.
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) *Client {
		c.Transport = t
		return c
	}
}

// WithCheckRedirect returns an Option that sets the client CheckRedirect
// function to the provided function.
func WithCheckRedirect(f func(req *http.Request, via []*http.Request) error) Option {
	return func(c *Client) *Client {
		c.CheckRedirect = f
		return c
	}
}

// WithJar returns an Option that sets the client's cookie jar to the provided
// value.
func WithJar(j http.CookieJar) Option {
	return func(c *Client) *Client {
		c.Jar = j
		return c
	}
}

// New returns a client, optionally modified by passing it through the given
// Option functions.
func New(opts ...Option) *Client {
	c := &Client{
		Client: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		c = opt(c)
	}

	return c
}
