package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/service"
)

const (
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultReadHeaderTimeout = 10 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, tlsConfig *tls.Config, authenticator auth.Authenticator, helloService service.HelloService) *Server {
	handler := NewHandler(authenticator, helloService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			TLSConfig:         tlsConfig,
			ReadTimeout:       defaultReadTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		if s.httpServer.TLSConfig != nil {
			slog.Info("Starting HTTPS server", "addr", s.httpServer.Addr)
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			slog.Info("Starting HTTP server", "addr", s.httpServer.Addr)
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, defaultShutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(shutdownCtx)
}
