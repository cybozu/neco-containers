package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type Metrics struct {
	*Config
}

func NewMetrics(cfg *Config) (*Metrics, error) {
	if cfg == nil {
		return nil, errors.New("metrics config is nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Metrics{cfg}, nil
}

// Serve runs the Prometheus metrics endpoint until ctx is cancelled or the
// underlying HTTP server returns an error. Returning nil means a graceful
// shutdown; any other value is a real server failure callers may choose to
// retry.
func (m *Metrics) Serve(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	srv := &http.Server{
		Addr:    m.AddrPort,
		Handler: mux,
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
			return
		}
		close(serveErr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-serveErr:
		return err
	}
}
