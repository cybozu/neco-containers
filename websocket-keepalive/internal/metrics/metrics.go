package metrics

import (
	"errors"
	"net/http"

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

func (m *Metrics) Serve() error {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	return http.ListenAndServe(m.AddrPort, nil)
}
