package watch

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
)

// Watcher watches target server health and creates metrics from it.
type Watcher struct {
	targetAddrs []string
	interval    time.Duration
	httpClient  *http.Client
}

// NewWatcher creates an Ingress watcher.
func NewWatcher(
	targetURLs []string,
	interval time.Duration,
	httpClient *http.Client,
) *Watcher {
	return &Watcher{
		targetAddrs: targetURLs,
		interval:    interval,
		httpClient:  httpClient,
	}
}

// Run repeats to get server health and send it via channel.
func (w *Watcher) Run(ctx context.Context) error {
	env := well.NewEnvironment(ctx)
	for _, t := range w.targetAddrs {
		t := t
		env.Go(func(ctx context.Context) error {
			tick := time.NewTicker(w.interval)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-tick.C:
					res, err := w.httpClient.Get(t)
					if err != nil {
						log.Info("GET failed.", map[string]interface{}{
							"url":       t,
							log.FnError: err,
						})
						metrics.HTTPGetFailTotal.WithLabelValues(t).Inc()
					} else {
						log.Info("GET succeeded.", map[string]interface{}{
							"url": t,
						})
						metrics.HTTPGetSuccessfulTotal.WithLabelValues(res.Status, t).Inc()
						res.Body.Close()
					}
				}
			}
		})
	}
	env.Stop()
	return env.Wait()
}
