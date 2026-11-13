package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// proxy is a reverse proxy for BMCs.
// It holds the TLS certificate and the map from hostnames to BMC addresses,
// which are updated periodically and read concurrently by the tunnels.
type proxy struct {
	// useTLS is false if the proxy serves plain HTTP.
	useTLS bool
	// readHeaderTimeout is the maximum duration for reading the request headers from clients.
	// The request body is not limited, so that uploading large files such as firmware images works.
	readHeaderTimeout time.Duration
	// idleTimeout is the maximum duration to wait for the next request on a keep-alive connection.
	idleTimeout time.Duration

	certificate atomic.Pointer[tls.Certificate]
	resolveMap  atomic.Pointer[map[string]string]
}

// updateCertificate replaces the TLS certificate served to clients.
func (p *proxy) updateCertificate(cert tls.Certificate) {
	p.certificate.Store(&cert)
}

// updateResolveMap replaces the map from hostnames to BMC addresses.
func (p *proxy) updateResolveMap(resolveMap map[string]string) {
	p.resolveMap.Store(&resolveMap)
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

// lowercaseKeys returns a copy of the resolve map with the keys in lower case,
// so that it can be looked up case-insensitively, because hostnames are case-insensitive.
// The resolve map has hostnames in lower case and serials in upper case.
func lowercaseKeys(resolveMap map[string]string) map[string]string {
	m := make(map[string]string, len(resolveMap))
	for k, v := range resolveMap {
		m[strings.ToLower(k)] = v
	}
	return m
}

// getHostname returns the key to look up the resolve map from the host of a request,
// which may include a port like "rack1-cs1.bmc.example.com:5900".
// The key is the first label of the host in lower case. See lowercaseKeys.
func getHostname(host string) string {
	// The Host header includes the port if it is not the default one.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	hostname, _, _ := strings.Cut(host, ".")
	return strings.ToLower(hostname)
}

// lookupBMCAddress returns the BMC address for the host of a request.
// resolveMap must be converted by lowercaseKeys.
// It returns false if the host is not found in the resolve map.
func lookupBMCAddress(host string, resolveMap map[string]string) (string, bool) {
	address, ok := resolveMap[getHostname(host)]
	return address, ok
}

// backendKey is the context key for the address of the BMC, in the form host:port.
type backendKey struct{}

// tunnelHandler returns the handler that forwards requests to the inner port of the BMC
// resolved from the host of each request. The connection to the BMC always uses HTTPS.
func (p *proxy) tunnelHandler(inner uint16) http.Handler {
	rewrite := func(pr *httputil.ProxyRequest) {
		pr.Out.URL.Host = pr.In.Context().Value(backendKey{}).(string)
		pr.Out.URL.Scheme = "https"
		pr.SetXForwarded()
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// errorHandler replaces the default one, which logs at INFO level via the log package.
	// r is the outbound request in most cases, but it is the inbound request for some errors,
	// so the address of the BMC is taken from the context instead of r.URL.
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
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

	// Resolve the BMC address before proxying, so that requests for unknown hosts are not forwarded anywhere.
	// They get 404 instead of 502, because there is no BMC for the host rather than a failure of a BMC.
	// The log level is WARN, because it is an error of the client such as a typo or a scan from the internet.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address, ok := lookupBMCAddress(r.Host, p.getResolveMap())
		if !ok {
			slog.Warn("failed to resolve hostname", "http_host", r.Host, "hostname", getHostname(r.Host))
			http.Error(w, "unknown host", http.StatusNotFound)
			return
		}
		backend := net.JoinHostPort(address, strconv.Itoa(int(inner)))
		rp.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), backendKey{}, backend)))
	})

	return accessLog(handler)
}

// runTunnel serves tunnelHandler on the external port.
// If p.useTLS is false, it serves plain HTTP.
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
	slog.Info("listening", "addr", ln.Addr().String(), "inner_port", inner, "tls", p.useTLS)

	if p.useTLS {
		ln = tls.NewListener(ln, &tls.Config{
			NextProtos:     []string{"h2", "http/1.1"},
			GetCertificate: p.getCertificate,
		})
	}

	return serve(ctx, server, ln)
}
