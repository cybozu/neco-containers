package client

import (
	internalmetrics "github.com/neco-containers/tcp-keepalive/internal/metrics"

	"github.com/VictoriaMetrics/metrics"
)

var (
	retryTotal             *metrics.Counter
	retryCount             *metrics.Counter
	sendSuccessTotal       *metrics.Counter
	sendErrorTotal         *metrics.Counter
	sendTimeoutTotal       *metrics.Counter
	receiveSuccessTotal    *metrics.Counter
	receiveErrorTotal      *metrics.Counter
	receiveTimeoutTotal    *metrics.Counter
	connStateUnestablished *metrics.Counter
	connStateEstablished   *metrics.Counter
	connStateClosed        *metrics.Counter
)

func initMetrics() {
	retryTotal = metrics.NewCounter(`retry_total{role="client"}`)
	retryCount = metrics.NewCounter(`retry_count{role="client"}`)
	sendSuccessTotal = metrics.NewCounter(`send_total{role="client",result="success"}`)
	sendErrorTotal = metrics.NewCounter(`send_total{role="client",result="error"}`)
	sendTimeoutTotal = metrics.NewCounter(`send_total{role="client",result="timeout"}`)
	receiveSuccessTotal = metrics.NewCounter(`receive_total{role="client",result="success"}`)
	receiveErrorTotal = metrics.NewCounter(`receive_total{role="client",result="error"}`)
	receiveTimeoutTotal = metrics.NewCounter(`receive_total{role="client",result="timeout"}`)
	connStateUnestablished = metrics.NewCounter(`connection{role="client",state="unestablished"}`)
	connStateEstablished = metrics.NewCounter(`connection{role="client",state="established"}`)
	connStateClosed = metrics.NewCounter(`connection{role="client",state="closed"}`)
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

func (m *Metrics) setConnStateUnestablished() {
	connStateUnestablished.Set(1)
	connStateEstablished.Set(0)
	connStateClosed.Set(0)
}

func (m *Metrics) setConnStateEstablished() {
	connStateUnestablished.Set(0)
	connStateEstablished.Set(1)
	connStateClosed.Set(0)
}

func (m *Metrics) setConnStateClosed() {
	connStateUnestablished.Set(0)
	connStateEstablished.Set(0)
	connStateClosed.Set(1)
}
