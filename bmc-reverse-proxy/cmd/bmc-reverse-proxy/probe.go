package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// probeHandler returns the handler for the probes of Kubernetes.
//
//   - /healthz always returns 200 while the process is running.
//   - /readyz returns 200 after the resolve map and the TLS certificate (if TLS is used) have been loaded,
//     and 503 before that.
func (p *proxy) probeHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok") //nolint:errcheck
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if p.getResolveMap() == nil {
			http.Error(w, "resolve map is not loaded", http.StatusServiceUnavailable)
			return
		}
		if p.useTLS {
			if _, err := p.getCertificate(nil); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		fmt.Fprintln(w, "ok") //nolint:errcheck
	})
	return mux
}

// runProbeServer serves the probe endpoints on the port.
// It uses a port separate from the tunnels because all requests to the tunnels are forwarded to BMCs.
func (p *proxy) runProbeServer(ctx context.Context, port uint16) error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           p.probeHandler(),
		ReadHeaderTimeout: p.readHeaderTimeout,
		IdleTimeout:       p.idleTimeout,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %w", server.Addr, err)
	}
	slog.Info("listening for probes", "addr", ln.Addr().String())

	return serve(ctx, server, ln)
}
