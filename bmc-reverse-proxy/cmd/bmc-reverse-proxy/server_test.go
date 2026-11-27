package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestServeWaitsForShutdown checks that serve returns after active requests finish.
func TestServeWaitsForShutdown(t *testing.T) {
	requestStarted := make(chan struct{})
	finishRequest := make(chan struct{})
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(requestStarted)
			<-finishRequest
		}),
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, server, ln) }()

	// Start a request, and stop serve while it is active.
	go http.Get("http://" + ln.Addr().String()) //nolint:errcheck
	<-requestStarted
	cancel()

	expectNotReturned(t, served)
	close(finishRequest)
	expectReturned(t, served)
}

// TestServeWaitsForHijackedConnections checks that serve waits for hijacked connections.
// http.Server.Shutdown does not wait for them.
// Like httputil.ReverseProxy, the handler returns after it finishes using the connection.
func TestServeWaitsForHijackedConnections(t *testing.T) {
	hijacked := make(chan struct{})
	finishRelay := make(chan struct{})
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, _, err := http.NewResponseController(w).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close() //nolint:errcheck
			close(hijacked)
			<-finishRelay
		}),
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, server, ln) }()

	// Hijack a connection, and stop serve while it is in use.
	// Send a raw request instead of using http.Get.
	// The handler does not respond after the hijack, so http.Get would block waiting for the response.
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close() //nolint:errcheck
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	<-hijacked
	cancel()

	expectNotReturned(t, served)
	close(finishRelay)
	expectReturned(t, served)
}

// expectNotReturned checks that serve does not return for a while.
func expectNotReturned(t *testing.T, served <-chan error) {
	t.Helper()
	select {
	case err := <-served:
		t.Fatalf("serve returned before the handler finished: %v", err)
	case <-time.After(200 * time.Millisecond):
		// OK: serve is still waiting for the handler.
		return
	}
}

// expectReturned checks that serve returns without an error.
func expectReturned(t *testing.T, served <-chan error) {
	t.Helper()
	select {
	case err := <-served:
		if err != nil {
			t.Error(err)
			return
		}
		// OK: serve returns without an error.
		return
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after the handler finished")
	}
}
