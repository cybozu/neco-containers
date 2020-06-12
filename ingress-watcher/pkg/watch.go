package pkg

import (
	"context"
	"net/http"
	"time"
)

// Watcher watches target server health and creates metrics from it.
type Watcher struct {
	endpoint   string
	interval   time.Duration
	channel    chan string
	httpClient *http.Client
}

// NewWatcher creates an Ingress watcher.
func NewWatcher(
	endpoint string,
	interval time.Duration,
	channel chan string,
	httpClient *http.Client,
) *Watcher {
	return &Watcher{
		endpoint:   endpoint,
		interval:   interval,
		channel:    channel,
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
			health, err := w.httpClient.Get(w.endpoint)
			if err != nil {
				return err
			}
			w.channel <- health.Status
		}
	}
}
