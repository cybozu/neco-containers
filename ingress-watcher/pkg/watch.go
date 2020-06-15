package pkg

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
)

// Watcher watches target server health and creates metrics from it.
type Watcher struct {
	tagetAddrs []string
	interval   time.Duration
	httpClient *http.Client
}

// NewWatcher creates an Ingress watcher.
func NewWatcher(
	targetAddrs []string,
	interval time.Duration,
	httpClient *http.Client,
) *Watcher {
	return &Watcher{
		tagetAddrs: targetAddrs,
		interval:   interval,
		httpClient: httpClient,
	}
}

// Run repeats to get server health and send it via channel.
func (w *Watcher) Run(ctx context.Context) error {
	env := well.NewEnvironment(ctx)
	for _, t := range w.tagetAddrs {
		t := t
		env.Go(func(ctx context.Context) error {
			tick := time.NewTicker(w.interval)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-tick.C:
					res, err := w.httpClient.Get("http://" + t)
					if err != nil {
						metrics.HTTPGetFailTotal.WithLabelValues(t).Inc()
					} else {
						metrics.HTTPGetSuccessfulTotal.WithLabelValues(res.Status, t).Inc()
						res.Body.Close()
					}

					res, err = w.httpClient.Get("https://" + t)
					if err != nil {
						metrics.HTTPSGetFailTotal.WithLabelValues(t).Inc()
					} else {
						metrics.HTTPSGetSuccessfulTotal.WithLabelValues(res.Status, t).Inc()
						res.Body.Close()
					}
				}
			}
		})
	}
	env.Stop()
	return env.Wait()
}
