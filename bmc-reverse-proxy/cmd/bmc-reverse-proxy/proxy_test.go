package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestUpdateResolveMap checks that the keys are converted to lower case.
// It also checks the result that tells whether the map is changed.
func TestUpdateResolveMap(t *testing.T) {
	p := &proxy{}
	if p.getResolveMap() != nil {
		t.Fatal("the resolve map should not be loaded yet")
	}

	m := map[string]string{"ABC1234": "192.0.2.4", "XYZ5678": "192.0.2.5"}
	if !p.updateResolveMap(m) {
		t.Error("the first update should be a change")
	}
	if expected := map[string]string{"abc1234": "192.0.2.4", "xyz5678": "192.0.2.5"}; !maps.Equal(p.getResolveMap(), expected) {
		t.Errorf("expected %v, got %v", expected, p.getResolveMap())
	}

	if p.updateResolveMap(maps.Clone(m)) {
		t.Error("the same content should not be a change")
	}
	m["ABC1234"] = "192.0.2.6"
	if !p.updateResolveMap(m) {
		t.Error("a different address should be a change")
	}

	// An empty map is regarded as loaded.
	p = &proxy{}
	if !p.updateResolveMap(map[string]string{}) || p.getResolveMap() == nil {
		t.Error("an empty map should be loaded")
	}
}

func TestLookupKey(t *testing.T) {
	cases := []struct {
		name     string
		host     string
		expected string
	}{
		{name: "LowerCase", host: "abc1234.bmc.example.com", expected: "abc1234"},
		{name: "UpperCase", host: "ABC1234.bmc.example.com", expected: "abc1234"},
		{name: "MixedCase", host: "Xyz5678.bmc.example.com", expected: "xyz5678"},
		{name: "WithPort", host: "abc1234.bmc.example.com:5900", expected: "abc1234"},
		{name: "SingleLabelWithPort", host: "localhost:8443", expected: "localhost"},
		// Only the first label is used.
		{name: "SerialNotInFirstLabel", host: "bmc.abc1234.example.com", expected: "bmc"},
		{name: "Empty", host: "", expected: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if key := lookupKey(c.host); key != c.expected {
				t.Errorf("expected %q, got %q", c.expected, key)
			}
		})
	}
}

func TestLookupBMCAddress(t *testing.T) {
	// The keys are serials in upper case, as written by update-resolve-map.
	p := &proxy{}
	p.updateResolveMap(map[string]string{
		"ABC1234": "192.0.2.1",
		"XYZ5678": "192.0.2.2",
	})

	cases := []struct {
		name     string
		key      string
		expected string // empty if not found
	}{
		{name: "Found", key: "abc1234", expected: "192.0.2.1"},
		{name: "FoundAnother", key: "xyz5678", expected: "192.0.2.2"},
		{name: "NotExist", key: "notexist"},
		{name: "Empty", key: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			address, found := p.lookupBMCAddress(c.key)
			if address != c.expected || found != (c.expected != "") {
				t.Errorf("expected %q, got (%q, %v)", c.expected, address, found)
			}
		})
	}

	t.Run("NotLoaded", func(t *testing.T) {
		if _, found := (&proxy{}).lookupBMCAddress("abc1234"); found {
			t.Error("nothing should be found before the resolve map is loaded")
		}
	})
}

// TestTunnel checks that requests are forwarded to the BMC address in the resolve map.
// Requests for unknown hosts must not be forwarded.
func TestTunnel(t *testing.T) {
	var backendRequests atomic.Int32
	inner := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendRequests.Add(1)
		fmt.Fprintf(w, "%s %s", r.Host, r.URL.RequestURI()) //nolint:errcheck
	}))
	tunnel := startTunnel(t, inner)

	get := func(host string) (int, string) {
		t.Helper()
		// The query contains ";", which ReverseProxy removes by default.
		req, err := http.NewRequest(http.MethodGet, tunnel.URL+"/foo?a=1;b=2", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = host
		resp, err := tunnel.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() //nolint:errcheck
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, string(body)
	}

	status, body := get("abc1234.bmc.example.com")
	if status != http.StatusOK || body != "abc1234.bmc.example.com /foo?a=1;b=2" {
		t.Errorf("unexpected response for a known host: %d %q", status, body)
	}

	status, _ = get("unknown.bmc.example.com")
	if status != http.StatusNotFound {
		t.Errorf("unexpected status for an unknown host: %d", status)
	}
	if n := backendRequests.Load(); n != 1 {
		t.Errorf("the backend should receive only the request for the known host, got %d", n)
	}
}

// TestTunnelUpgrade checks that a protocol upgrade, e.g. to WebSocket, is proxied.
// It also checks that the access log records 101.
// ReverseProxy writes the 101 response to the hijacked connection, not via WriteHeader.
func TestTunnelUpgrade(t *testing.T) {
	inner := startBackend(t, http.HandlerFunc(echoProtocolHandler))
	logBuf := captureLog(t)
	tunnel := startTunnel(t, inner)

	// Switch to the echo protocol through the tunnel.
	conn, err := net.Dial("tcp", tunnel.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close() //nolint:errcheck
	br := bufio.NewReader(conn)
	// Send a raw request instead of using http.Client.
	// After the upgrade, the same connection is used for the echo protocol, which is not HTTP.
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: abc1234.bmc.example.com\r\n" + upgradeToEchoHeaders)); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", resp.StatusCode)
	}

	// Data is relayed in both directions.
	if _, err := conn.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	if line, err := br.ReadString('\n'); err != nil || line != "hello\n" {
		t.Errorf("unexpected echo: %q, %v", line, err)
	}

	// The access log is written after the upgraded connection is closed.
	conn.Close() //nolint:errcheck
	if status := waitAccessLogStatus(logBuf); status != http.StatusSwitchingProtocols {
		t.Errorf("expected the access log with status 101, got %d\n%s", status, logBuf.Bytes())
	}
}

// TestTunnelClientCanceled checks that a request canceled by the client is recorded with 499.
// The proxy must not log it as an error.
func TestTunnelClientCanceled(t *testing.T) {
	received := make(chan struct{})
	inner := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(received)
		<-r.Context().Done()
	}))
	logBuf := captureLog(t)
	tunnel := startTunnel(t, inner)

	// Cancel the request after the backend receives it.
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		<-received
		cancel()
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tunnel.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "abc1234.bmc.example.com"
	if _, err := tunnel.Client().Do(req); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if status := waitAccessLogStatus(logBuf); status != statusClientClosedRequest {
		t.Errorf("expected the access log with status %d, got %d\n%s", statusClientClosedRequest, status, logBuf.Bytes())
	}
	if bytes.Contains(logBuf.Bytes(), []byte(`"msg":"proxy error"`)) {
		t.Errorf("a request canceled by the client should not be logged as a proxy error\n%s", logBuf.Bytes())
	}
}

// Helpers for the tests of the tunnel.

// startBackend starts a TLS server as a BMC, and returns its port.
func startBackend(t *testing.T, handler http.Handler) uint16 {
	t.Helper()
	backend := httptest.NewTLSServer(handler)
	t.Cleanup(backend.Close)
	return uint16(backend.Listener.Addr().(*net.TCPAddr).Port)
}

// startTunnel starts a tunnel to the backend port.
// Its resolve map points the serial ABC1234 to the local host.
func startTunnel(t *testing.T, inner uint16) *httptest.Server {
	t.Helper()
	p := &proxy{}
	p.updateResolveMap(map[string]string{"ABC1234": "127.0.0.1"})
	tunnel := httptest.NewServer(p.tunnelHandler(inner))
	t.Cleanup(tunnel.Close)
	return tunnel
}

// upgradeToEchoHeaders is the end of the headers to switch to "echo".
// "echo" is a protocol only for the tests.
const upgradeToEchoHeaders = "Connection: Upgrade\r\nUpgrade: echo\r\n\r\n"

// echoProtocolHandler switches to the echo protocol.
// It is like a BMC switching to WebSocket for the virtual console.
// It takes over the connection by Hijack, and sends back a line it receives.
func echoProtocolHandler(w http.ResponseWriter, r *http.Request) {
	conn, brw, err := http.NewResponseController(w).Hijack()
	if err != nil {
		return
	}
	defer conn.Close() //nolint:errcheck

	_, _ = brw.WriteString("HTTP/1.1 101 Switching Protocols\r\n" + upgradeToEchoHeaders)
	_ = brw.Flush()
	line, _ := brw.ReadString('\n')
	_, _ = brw.WriteString(line)
	_ = brw.Flush()
}

// captureLog makes slog write JSON logs to the returned buffer during the test.
func captureLog(t *testing.T) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
	return buf
}

// waitAccessLogStatus waits for an access log in the buffer, and returns its status code.
// It returns 0 if not found.
// The access log is written after the handler returns, e.g. after an upgraded connection is closed.
func waitAccessLogStatus(buf *syncBuffer) int {
	for range 100 {
		for line := range bytes.SplitSeq(buf.Bytes(), []byte("\n")) {
			var entry struct {
				Msg    string `json:"msg"`
				Status int    `json:"http_status_code"`
			}
			if json.Unmarshal(line, &entry) == nil && entry.Msg == "access" {
				return entry.Status
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return 0
}

// syncBuffer is a bytes.Buffer safe for concurrent writes from the logger.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buf.Bytes())
}
