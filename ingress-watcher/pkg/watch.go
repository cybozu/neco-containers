package pkg

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
)

// Watcher watches target server health and creates metrics from it.
type Watcher struct {
	endpoint   string
	interval   time.Duration
	httpClient *http.Client
}

// NewWatcher creates an Ingress watcher.
func NewWatcher(
	endpoint string,
	interval time.Duration,
	httpClient *http.Client,
) *Watcher {
	return &Watcher{
		endpoint:   endpoint,
		interval:   interval,
		httpClient: httpClient,
	}
}

// Run repeats to get server health and send it via channel.
func (w *Watcher) Run(ctx context.Context) error {
	tick := time.NewTicker(w.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			res, err := w.httpClient.Get("http://" + w.endpoint)
			if err != nil {
				metrics.HTTPGetFailTotal.WithLabelValues(w.endpoint).Inc()
			} else {
				metrics.HTTPGetSuccessfulTotal.WithLabelValues(res.Status, w.endpoint).Inc()
			}

			res, err = w.httpClient.Get("https://" + w.endpoint)
			if err != nil {
				metrics.HTTPSGetFailTotal.WithLabelValues(w.endpoint).Inc()
			} else {
				metrics.HTTPSGetSuccessfulTotal.WithLabelValues(res.Status, w.endpoint).Inc()
			}
		}
	}
}
