package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

const (
	reloadInterval = 60 * time.Second
	// retryInterval is used instead of reloadInterval until the resolve map is loaded,
	// so that the Pod becomes ready soon after update-resolve-map writes the file.
	retryInterval = 1 * time.Second
)

// Ports of the tunnels. Each tunnel listens on the external port and forwards requests to the BMC port.
const (
	// Web interface of the BMC.
	webBMCPort      uint16 = 443
	webExternalPort uint16 = 8443

	// Virtual console of the BMC.
	virtualConsoleBMCPort      uint16 = 5900
	virtualConsoleExternalPort uint16 = 5900
)

// probePort is the port for the probes of Kubernetes (/healthz and /readyz).
const probePort uint16 = 8080

// Timeouts of the HTTP servers for clients.
const (
	readHeaderTimeout = 30 * time.Second
	idleTimeout       = 30 * time.Second
)

var (
	flagNoTLS      = flag.Bool("no-tls", false, "serve plain HTTP without loading the TLS certificate (for local testing only)")
	flagTLSCert    = flag.String("tls-cert", "/etc/tls/tls.crt", "path of the TLS certificate file")
	flagTLSKey     = flag.String("tls-key", "/etc/tls/tls.key", "path of the TLS private key file")
	flagResolveMap = flag.String("resolve-map", resolvemap.DefaultPath, "path of the resolve map file written by update-resolve-map")
)

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	p := &proxy{
		useTLS:            !*flagNoTLS,
		readHeaderTimeout: readHeaderTimeout,
		idleTimeout:       idleTimeout,
	}
	slog.Info("starting", "resolve_map", *flagResolveMap, "tls", p.useTLS)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	g, gctx := errgroup.WithContext(ctx)
	// gctx is canceled by a signal or an error of any goroutine. Its cause tells which.
	context.AfterFunc(gctx, func() {
		slog.Info("shutting down", "reason", context.Cause(gctx).Error())
	})

	// Reload the TLS certificate periodically.
	// Exit if it cannot be loaded, because the proxy cannot serve TLS without it.
	if p.useTLS {
		g.Go(func() error {
			for {
				cert, err := tls.LoadX509KeyPair(*flagTLSCert, *flagTLSKey)
				if err != nil {
					return fmt.Errorf("failed to load x509 key pair from (%s, %s): %w", *flagTLSCert, *flagTLSKey, err)
				}
				p.updateCertificate(cert)

				select {
				case <-gctx.Done():
					return nil
				case <-time.After(reloadInterval):
				}
			}
		})
	}

	// Reload the resolve map file periodically.
	// The file may not exist yet just after the Pod starts, or may be broken temporarily.
	// Keep running with the last loaded map in such cases.
	g.Go(func() error {
		for {
			m, err := resolvemap.Read(*flagResolveMap)
			if err == nil {
				// Look up the map case-insensitively, because hostnames are case-insensitive.
				m = lowercaseKeys(m)
			}
			prev := p.getResolveMap()
			switch {
			case errors.Is(err, fs.ErrNotExist) && prev == nil:
				// update-resolve-map has not written the file yet just after the Pod starts.
				slog.Info("waiting for resolve map", "path", *flagResolveMap)
			case err != nil:
				slog.Error("failed to read resolve map", "path", *flagResolveMap, "error", err)
			case prev == nil || !maps.Equal(prev, m):
				// Log only when the content has changed, including the first load.
				slog.Info("resolve map changed", "path", *flagResolveMap, "entries", len(m))
				p.updateResolveMap(m)
			}

			interval := reloadInterval
			if p.getResolveMap() == nil {
				interval = retryInterval
			}
			select {
			case <-gctx.Done():
				return nil
			case <-time.After(interval):
			}
		}
	})

	g.Go(func() error {
		return p.runTunnel(gctx, webBMCPort, webExternalPort)
	})

	g.Go(func() error {
		return p.runTunnel(gctx, virtualConsoleBMCPort, virtualConsoleExternalPort)
	})

	g.Go(func() error {
		return p.runProbeServer(gctx, probePort)
	})

	// The goroutines return nil when they are canceled, so an error here is a real failure.
	if err := g.Wait(); err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
	slog.Info("stopped")
}
