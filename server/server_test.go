package server

import (
	"testing"
	"time"
)

func TestServerOptions(t *testing.T) {
	s := New(":8080", nil)
	if s.server.ReadTimeout != ReadTimeout {
		t.Errorf("expected read timeout %s, got %s", ReadTimeout, s.server.ReadTimeout)
	}

	to := 10 * time.Second
	s2 := New(":8080", nil, WithReadTimeout(to))
	if s2.server.ReadTimeout != to {
		t.Errorf("expected read timeout %s, got %s", to, s.server.ReadTimeout)
	}
}
