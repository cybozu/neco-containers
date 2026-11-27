package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// proxy is a reverse proxy for BMCs.
// It holds the TLS certificate and the map from serials to BMC addresses.
// They are updated periodically and read by the tunnels concurrently.
type proxy struct {
	// readHeaderTimeout limits the time to read the request headers.
	// The request body is not limited, to allow large uploads such as firmware images.
	readHeaderTimeout time.Duration
	// idleTimeout limits the time to wait for the next request on a keep-alive connection.
	idleTimeout time.Duration

	certificate atomic.Pointer[tls.Certificate]
	resolveMap  atomic.Pointer[map[string]string]
}

// updateCertificate replaces the TLS certificate served to clients.
func (p *proxy) updateCertificate(cert tls.Certificate) {
	p.certificate.Store(&cert)
}

// updateResolveMap replaces the map from serials to BMC addresses.
// It returns true if the map is changed.
// The keys are converted to lower case, because hostnames are case-insensitive.
func (p *proxy) updateResolveMap(resolveMap map[string]string) bool {
	m := make(map[string]string, len(resolveMap))
	for k, v := range resolveMap {
		m[strings.ToLower(k)] = v
	}
	if prev := p.resolveMap.Load(); prev != nil && maps.Equal(*prev, m) {
		return false
	}
	p.resolveMap.Store(&m)
	return true
}

// getCertificate is used as tls.Config.GetCertificate.
func (p *proxy) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert := p.certificate.Load()
	if cert == nil {
		return nil, errors.New("certificate is not loaded")
	}
	return cert, nil
}

// getResolveMap returns nil if the resolve map has not been loaded yet.
func (p *proxy) getResolveMap() map[string]string {
	m := p.resolveMap.Load()
	if m == nil {
		return nil
	}
	return *m
}

// lookupKey returns the key to look up the resolve map from the host of a request.
// The key is the first label of the host in lower case.
// e.g. "abc1234" for "ABC1234.bmc.example.com:5900".
func lookupKey(host string) string {
	// The Host header includes the port if it is not the default one.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	label, _, _ := strings.Cut(host, ".")
	return strings.ToLower(label)
}

// lookupBMCAddress returns the BMC address for the key returned by lookupKey.
// It returns false if the key is not found in the resolve map.
func (p *proxy) lookupBMCAddress(key string) (string, bool) {
	address, ok := p.getResolveMap()[key]
	return address, ok
}

// statusClientClosedRequest is recorded in the access log when the client closes the connection.
// It is a convention of nginx.
const statusClientClosedRequest = 499

// backendKey is the context key for the address of the BMC (host:port).
type backendKey struct{}

// tunnelHandler returns the handler that forwards requests to the inner port of the BMC.
// The BMC is resolved from the host of each request. The connection to the BMC uses HTTPS.
func (p *proxy) tunnelHandler(inner uint16) http.Handler {
	rewrite := func(pr *httputil.ProxyRequest) {
		pr.Out.URL.Host = pr.In.Context().Value(backendKey{}).(string)
		pr.Out.URL.Scheme = "https"
		// ReverseProxy removes unparsable query parameters, e.g. ones containing ";".
		// Forward the query as is.
		pr.Out.URL.RawQuery = pr.In.URL.RawQuery
		pr.SetXForwarded()
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Do not use HTTP proxies. BMCs are in the internal network.
	transport.Proxy = nil
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// errorHandler replaces the default one, which logs at INFO level.
	// r may be the inbound request for some errors.
	// So the address of the BMC is taken from the context, not from r.URL.
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		// The client closed the connection, e.g. closed the browser tab.
		// It is not an error of the proxy or the BMC. Only record it in the access log.
		if errors.Is(err, context.Canceled) {
			w.WriteHeader(statusClientClosedRequest)
			return
		}
		backend, _ := r.Context().Value(backendKey{}).(string)
		slog.Error("proxy error",
			"error", err,
			"http_host", r.Host,
			"url", r.RequestURI,
			"backend", backend,
		)
		w.WriteHeader(http.StatusBadGateway)
	}

	rp := &httputil.ReverseProxy{
		Rewrite:      rewrite,
		Transport:    transport,
		ErrorHandler: errorHandler,
	}

	// Resolve the BMC address before proxying. Requests for unknown hosts are not forwarded.
	// They get 404, not 502, because no BMC is failing.
	// The log level is WARN, because it is an error of the client, e.g. a typo or a scan.
	port := strconv.Itoa(int(inner))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := lookupKey(r.Host)
		address, ok := p.lookupBMCAddress(key)
		if !ok {
			slog.Warn("unknown host", "http_host", r.Host, "key", key)
			http.Error(w, "unknown host", http.StatusNotFound)
			return
		}
		backend := net.JoinHostPort(address, port)
		rp.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), backendKey{}, backend)))
	})

	return accessLog(handler)
}

// runTunnel serves tunnelHandler on the external port over TLS.
func (p *proxy) runTunnel(ctx context.Context, inner uint16, external uint16) error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", external),
		Handler:           p.tunnelHandler(inner),
		ReadHeaderTimeout: p.readHeaderTimeout,
		IdleTimeout:       p.idleTimeout,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %w", server.Addr, err)
	}
	slog.Info("listening", "addr", ln.Addr().String(), "inner_port", inner)

	ln = tls.NewListener(ln, &tls.Config{
		NextProtos:     []string{"h2", "http/1.1"},
		GetCertificate: p.getCertificate,
	})

	return serve(ctx, server, ln)
}
