package client_test

import (
	"testing"
	"time"

	"github.com/haleyrc/http/client"
)

func TestClientTimeout(t *testing.T) {
	c := client.New()
	if c.Timeout != 5*time.Second {
		t.Errorf("expected timeout %s, got %s", 5*time.Second, c.Timeout)
	}

	to := 10 * time.Second
	c2 := client.New(client.WithTimeout(to))
	if c2.Timeout != to {
		t.Errorf("expected timeout %s, got %s", to, c2.Timeout)
	}
}
