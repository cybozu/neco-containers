package client

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	internalmetrics "github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
)

type Metrics struct {
	*internalmetrics.Metrics
}

var (
	established    *metrics.Gauge
	pingRetryTotal *metrics.Counter
	pingTotal      *metrics.Counter
	pongTotal      *metrics.Counter
)

func initMetrics(local, remote string) {
	established = metrics.NewGauge(fmt.Sprintf(`established{role="client",local="%s",remote="%s"}`, local, remote), nil)
	pingRetryTotal = metrics.NewCounter(fmt.Sprintf(`ping_retry_count_total{local="%s",remote="%s"}`, local, remote))
	pingTotal = metrics.NewCounter(fmt.Sprintf(`sent_ping_total{role="client",local="%s",remote="%s"}`, local, remote))
	pongTotal = metrics.NewCounter(fmt.Sprintf(`received_pong_total{role="client",local="%s",remote="%s"}`, local, remote))
}

func NewMetrics(cfg *internalmetrics.Config) (*Metrics, error) {
	m, err := internalmetrics.NewMetrics(cfg)
	if err != nil {
		return nil, err
	}

	return &Metrics{Metrics: m}, nil
}

func (m *Metrics) setEstablished() {
	established.Set(1)
}

func (m *Metrics) setUnestablished() {
	established.Set(0)
	pingRetryTotal.Set(0)
	pingTotal.Set(0)
	pongTotal.Set(0)
}

func (m *Metrics) incrementRetryCount() {
	pingRetryTotal.Inc()
}

func (m *Metrics) incrementPingTotal() {
	pingTotal.Inc()
}

func (m *Metrics) incrementPongTotal() {
	pongTotal.Inc()
}
