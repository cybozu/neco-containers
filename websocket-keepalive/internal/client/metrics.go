package client

import (
	"github.com/VictoriaMetrics/metrics"
	internalmetrics "github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
)

type Metrics struct {
	*internalmetrics.Metrics
}

var (
	established    *metrics.Counter
	pingRetryTotal *metrics.Counter
)

func initMetrics() {
	established = metrics.NewCounter(`established{}`)
	pingRetryTotal = metrics.NewCounter(`ping_retry_count_total{}`)
}

func NewMetrics(cfg *internalmetrics.Config) (*Metrics, error) {
	m, err := internalmetrics.NewMetrics(cfg)
	if err != nil {
		return nil, err
	}

	initMetrics()

	return &Metrics{m}, nil
}

func (m *Metrics) setEstablished() {
	established.Set(1)
}

func (m *Metrics) setUnestablished() {
	established.Set(0)
	pingRetryTotal.Set(0)
}

func (m *Metrics) incrementRetryCount() {
	pingRetryTotal.Inc()
}
