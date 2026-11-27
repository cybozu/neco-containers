// update-resolve-map writes the resolve map file for bmc-reverse-proxy.
// The file maps serials of machines to BMC addresses, based on sabakan.
// It runs as a sidecar of bmc-reverse-proxy, and updates the file periodically.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

// retryInterval is used instead of -interval until the file is written for the first time.
const retryInterval = 5 * time.Second

var (
	flagSabakanURL = flag.String("sabakan-url", "http://localhost:10080/", "URL of sabakan's REST API")
	flagOutput     = flag.String("output", resolvemap.DefaultPath, "path of the resolve map file to write")
	flagInterval   = flag.Duration("interval", time.Minute, "interval to update the resolve map file; if 0, update it once and exit")
)

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	sabakanURL, err := url.Parse(*flagSabakanURL)
	if err != nil || (sabakanURL.Scheme != "http" && sabakanURL.Scheme != "https") || sabakanURL.Host == "" {
		slog.Error("invalid sabakan-url (expected http://host:port/)", "sabakan_url", *flagSabakanURL)
		os.Exit(1)
	}
	if *flagInterval < 0 {
		slog.Error("invalid interval", "interval", flagInterval.String())
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sabakan := newSabakanClient(sabakanURL)

	// Update the file once and exit.
	if *flagInterval == 0 {
		if err := runOnce(ctx, sabakan, *flagOutput); err != nil {
			slog.Error("failed to update resolve map", "error", err)
			os.Exit(1)
		}
		return
	}

	// Update the file every -interval until a signal is received.
	slog.Info("starting", "output", *flagOutput, "interval", flagInterval.String())
	written := false
loop:
	for {
		// Keep running on errors, e.g. a temporary failure of sabakan.
		// bmc-reverse-proxy uses the last written file in the meantime.
		switch err := runOnce(ctx, sabakan, *flagOutput); {
		case err == nil:
			written = true
		case !errors.Is(err, context.Canceled):
			slog.Error("failed to update resolve map", "error", err)
		}

		// bmc-reverse-proxy is not ready until the file is written.
		// So retry soon until the first success.
		interval := *flagInterval
		if !written {
			interval = min(retryInterval, interval)
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(interval):
		}
	}
	slog.Info("stopped", "reason", context.Cause(ctx).Error())
}

// runOnce gets the machines from sabakan, and updates the resolve map file at output.
func runOnce(ctx context.Context, sabakan sabakanClient, output string) error {
	machines, err := sabakan.getMachines(ctx)
	if err != nil {
		return err
	}
	return update(machines, output)
}
