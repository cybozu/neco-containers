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
					url := "http://" + t
					res, err := w.httpClient.Get(url)
					if err != nil {
						log.Info("GET failed.", map[string]interface{}{
							"url":       url,
							log.FnError: err,
						})
						metrics.HTTPGetFailTotal.WithLabelValues(t).Inc()
					} else {
						log.Info("GET succeeded.", map[string]interface{}{
							"url": url,
						})
						metrics.HTTPGetSuccessfulTotal.WithLabelValues(res.Status, t).Inc()
						res.Body.Close()
					}

					url = "https://" + t
					res, err = w.httpClient.Get(url)
					if err != nil {
						log.Info("GET failed.", map[string]interface{}{
							"url":       url,
							log.FnError: err,
						})
						metrics.HTTPSGetFailTotal.WithLabelValues(t).Inc()
					} else {
						log.Info("GET succeeded.", map[string]interface{}{
							"url": url,
						})
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
