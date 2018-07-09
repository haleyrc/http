// Package http contains subpackages that expose alternative interfaces to the
// default Go implementations. Notably, we set sane defaults for timeouts for
// both clients and servers to prevent hung connections. For the server, we also
// add some basic signal handling and clean shutdown.
package http
