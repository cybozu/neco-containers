package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// startServe starts serve with handler, and returns the address, the function to stop it, and the channel of its result.
func startServe(t *testing.T, handler http.Handler) (string, context.CancelFunc, <-chan error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, &http.Server{Handler: handler}, ln) }()
	return ln.Addr().String(), cancel, served
}

// checkServeBlocks checks that serve does not return until finish is called.
func checkServeBlocks(t *testing.T, stop context.CancelFunc, served <-chan error, finish func()) {
	t.Helper()
	stop()
	select {
	case err := <-served:
		t.Fatalf("serve returned before the handler finished: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	finish()
	select {
	case err := <-served:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after the handler finished")
	}
}

// TestServeWaitsForShutdown checks that serve returns after active requests finish.
func TestServeWaitsForShutdown(t *testing.T) {
	requestStarted := make(chan struct{})
	finishRequest := make(chan struct{})
	addr, stop, served := startServe(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-finishRequest
	}))

	go http.Get("http://" + addr) //nolint:errcheck
	<-requestStarted

	checkServeBlocks(t, stop, served, func() { close(finishRequest) })
}

// TestServeWaitsForHijackedConnections checks that serve waits for the handlers of hijacked connections,
// which http.Server.Shutdown does not wait for. Like httputil.ReverseProxy relaying a WebSocket,
// the handler does not return until it finishes using the hijacked connection.
func TestServeWaitsForHijackedConnections(t *testing.T) {
	hijacked := make(chan struct{})
	finishRelay := make(chan struct{})
	addr, stop, served := startServe(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close() //nolint:errcheck
		close(hijacked)
		<-finishRelay
	}))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close() //nolint:errcheck
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	<-hijacked

	checkServeBlocks(t, stop, served, func() { close(finishRelay) })
}
