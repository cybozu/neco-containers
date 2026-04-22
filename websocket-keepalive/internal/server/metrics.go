package server

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"

	internalmetrics "github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
)

type Metrics struct {
	*internalmetrics.Metrics
}

type serverMetricsSet struct {
	established *metrics.Gauge
	pingTotal   *metrics.Counter
	pongTotal   *metrics.Counter
}

func initServerMetrics(local, remote string) *serverMetricsSet {
	// GetOrCreate* is used so reconnects from the same remote (after the kernel
	// reuses the port) return the existing metric instead of panicking on double
	// registration.
	return &serverMetricsSet{
		established: metrics.GetOrCreateGauge(fmt.Sprintf(`established{role="server",local="%s",remote="%s"}`, local, remote), nil),
		pingTotal:   metrics.GetOrCreateCounter(fmt.Sprintf(`received_ping_total{role="server",local="%s",remote="%s"}`, local, remote)),
		pongTotal:   metrics.GetOrCreateCounter(fmt.Sprintf(`sent_pong_total{role="server",local="%s",remote="%s"}`, local, remote)),
	}
}

func (m *serverMetricsSet) setEstablished() {
	m.established.Set(1)
}

func (m *serverMetricsSet) setUnestablished() {
	m.established.Set(0)
	m.pingTotal.Set(0)
	m.pongTotal.Set(0)
}

func (m *serverMetricsSet) incrementPingTotal() {
	m.pingTotal.Inc()
}

func (m *serverMetricsSet) incrementPongTotal() {
	m.pongTotal.Inc()
}

func NewMetrics(cfg *internalmetrics.Config) (*Metrics, error) {
	m, err := internalmetrics.NewMetrics(cfg)
	if err != nil {
		return nil, err
	}

	return &Metrics{Metrics: m}, nil
}
