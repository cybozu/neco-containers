package pkg

import (
	"context"
	"net/http"
	"sync"
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
	mutex      *sync.Mutex
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
	tick := time.NewTicker(w.interval)
	defer tick.Stop()

	env := well.NewEnvironment(ctx)
	for _, t := range w.tagetAddrs {
		env.Go(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-tick.C:
					err := w.watchHTTP(t)
					if err != nil {
						log.Warn("watch http failed", map[string]interface{}{
							log.FnError: err,
						})
					}

					err = w.watchHTTPS(t)
					if err != nil {
						log.Warn("watch https failed", map[string]interface{}{
							log.FnError: err,
						})
					}
				}
			}
		})
	}
	env.Stop()
	return env.Wait()
}

func (w *Watcher) watchHTTP(addr string) error {
	res, err := w.httpClient.Get("http://" + addr)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if err != nil {
		metrics.HTTPGetFailTotal.WithLabelValues(addr).Inc()
	} else {
		metrics.HTTPGetSuccessfulTotal.WithLabelValues(res.Status, addr).Inc()
	}
	return res.Body.Close()
}

func (w *Watcher) watchHTTPS(addr string) error {
	res, err := w.httpClient.Get("https://" + addr)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if err != nil {
		metrics.HTTPSGetFailTotal.WithLabelValues(addr).Inc()
	} else {
		metrics.HTTPSGetSuccessfulTotal.WithLabelValues(res.Status, addr).Inc()
	}
	return res.Body.Close()
}
