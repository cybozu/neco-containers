package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLookupBMCAddress(t *testing.T) {
	cases := []struct {
		name       string
		host       string
		resolveMap map[string]string
		expected   string
		found      bool
	}{
		{
			name:       "ResolveIPAddressFromHost1",
			host:       "rack1-cs1.example.com",
			resolveMap: map[string]string{"rack1-cs1": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromLowerCaseHost",
			host:       "123abcd.bmc.example.com",
			resolveMap: map[string]string{"123ABCD": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromUpperCaseHost",
			host:       "123ABCD.bmc.example.com",
			resolveMap: map[string]string{"123ABCD": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromMixedCaseHost",
			host:       "Rack1-CS4.bmc.example.com",
			resolveMap: map[string]string{"rack1-cs4": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromBootServerHost",
			host:       "boot-0.bmc.example.com",
			resolveMap: map[string]string{"boot-0": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			// The keys are used as is. The cluster name in the hostname of a boot server is not trimmed.
			name:       "ResolveIPAddressFromBootServerHostWithClusterName",
			host:       "stage0-boot-0.bmc.example.com",
			resolveMap: map[string]string{"boot-0": "192.0.2.1"},
			found:      false,
		},
		{
			name:       "ResolveIPAddressFromHostWithPort",
			host:       "rack1-cs1.bmc.example.com:5900",
			resolveMap: map[string]string{"rack1-cs1": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromSingleLabelHostWithPort",
			host:       "localhost:8443",
			resolveMap: map[string]string{"localhost": "192.0.2.1"},
			expected:   "192.0.2.1",
			found:      true,
		},
		{
			name:       "ResolveIPAddressFromHostNotExist",
			host:       "notexist.example.com",
			resolveMap: map[string]string{"rack1-cs1": "192.0.2.1"},
			found:      false,
		},
		{
			name:       "ResolveIPAddressFromHostWithResolveMapHasNotSet",
			host:       "rack1-cs1.example.com",
			resolveMap: nil,
			found:      false,
		},
		{
			name:       "ResolveIPAddressFromEmptyHost",
			host:       "",
			resolveMap: map[string]string{"rack1-cs1": "192.0.2.1"},
			found:      false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			address, found := lookupBMCAddress(c.host, lowercaseKeys(c.resolveMap))
			if address != c.expected || found != c.found {
				t.Errorf("expected (%q, %v), got (%q, %v)", c.expected, c.found, address, found)
			}
		})
	}
}

// startBackend starts a TLS server as a BMC, and returns its port.
func startBackend(t *testing.T, handler http.Handler) uint16 {
	t.Helper()
	backend := httptest.NewTLSServer(handler)
	t.Cleanup(backend.Close)
	port := backend.Listener.Addr().(*net.TCPAddr).Port
	return uint16(port)
}

// newTestProxy returns a proxy whose resolve map points rack1-cs1 to the local host.
func newTestProxy() *proxy {
	p := &proxy{}
	p.updateResolveMap(lowercaseKeys(map[string]string{"rack1-cs1": "127.0.0.1"}))
	return p
}

// TestTunnel checks that requests are forwarded to the BMC address in the resolve map,
// and that requests for unknown hosts are not forwarded anywhere.
func TestTunnel(t *testing.T) {
	var backendRequests atomic.Int32
	inner := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendRequests.Add(1)
		fmt.Fprintf(w, "%s %s", r.Host, r.URL.Path) //nolint:errcheck
	}))
	tunnel := httptest.NewServer(newTestProxy().tunnelHandler(inner))
	defer tunnel.Close()

	get := func(host string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, tunnel.URL+"/foo", nil)
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

	status, body := get("rack1-cs1.bmc.example.com")
	if status != http.StatusOK || body != "rack1-cs1.bmc.example.com /foo" {
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

// TestTunnelUpgrade checks that a protocol upgrade such as WebSocket is proxied
// and logged with status 101, though ReverseProxy writes the 101 response to the hijacked connection.
func TestTunnelUpgrade(t *testing.T) {
	// The backend switches to an echo protocol.
	inner := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, brw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()                                                                                  //nolint:errcheck
		brw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: echo\r\n\r\n") //nolint:errcheck
		brw.Flush()                                                                                         //nolint:errcheck
		line, _ := brw.ReadString('\n')
		brw.WriteString(line) //nolint:errcheck
		brw.Flush()           //nolint:errcheck
	}))

	// capture the access log
	var logBuf syncBuffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuf, nil)))
	defer slog.SetDefault(orig)

	tunnel := httptest.NewServer(newTestProxy().tunnelHandler(inner))
	defer tunnel.Close()

	conn, err := net.Dial("tcp", tunnel.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(conn)
	req := "GET / HTTP/1.1\r\nHost: rack1-cs1.bmc.example.com\r\nConnection: Upgrade\r\nUpgrade: echo\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", resp.StatusCode)
	}
	if _, err := conn.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	if line, err := br.ReadString('\n'); err != nil || line != "hello\n" {
		t.Errorf("unexpected echo: %q, %v", line, err)
	}
	conn.Close() //nolint:errcheck

	// The access log is written after the upgraded connection is closed.
	accessStatus := func() int {
		for line := range bytes.SplitSeq(logBuf.Bytes(), []byte("\n")) {
			var entry struct {
				Msg    string `json:"msg"`
				Status int    `json:"http_status_code"`
			}
			if json.Unmarshal(line, &entry) == nil && entry.Msg == "access" {
				return entry.Status
			}
		}
		return 0
	}
	var status int
	for range 100 {
		if status = accessStatus(); status != 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if status != http.StatusSwitchingProtocols {
		t.Errorf("expected the access log with status 101, got %d\n%s", status, logBuf.Bytes())
	}
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
