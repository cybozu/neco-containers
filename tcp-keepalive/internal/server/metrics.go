package server

import (
	internalmetrics "github.com/neco-containers/tcp-keepalive/internal/metrics"

	"github.com/VictoriaMetrics/metrics"
)

var (
	receiveSuccessTotal *metrics.Counter
	receiveErrorTotal   *metrics.Counter
	sendSuccessTotal    *metrics.Counter
	sendErrorTotal      *metrics.Counter
)

func initMetrics() {
	receiveSuccessTotal = metrics.NewCounter(`receive_total{role="server",result="success"}`)
	receiveErrorTotal = metrics.NewCounter(`receive_total{role="server",result="error"}`)
	sendSuccessTotal = metrics.NewCounter(`send_total{role="server",result="success"}`)
	sendErrorTotal = metrics.NewCounter(`send_total{role="server",result="error"}`)
}

type Metrics struct {
	*internalmetrics.Metrics
}

func NewMetrics(cfg *internalmetrics.Config) (*Metrics, error) {
	m, err := internalmetrics.NewMetrics(cfg)
	if err != nil {
		return nil, err
	}
	initMetrics()
	return &Metrics{m}, nil
}
