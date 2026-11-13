// update-resolve-map writes the resolve map file that bmc-reverse-proxy uses
// to resolve hostnames to BMC addresses, based on the machines registered with sabakan.
// It is expected to run as a sidecar of bmc-reverse-proxy and update the file periodically.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

// retryInterval is used instead of -interval until the resolve map file is written for the first time.
const retryInterval = 5 * time.Second

var (
	flagSabakanAddress = flag.String("sabakan-address", "", "address of sabakan's GraphQL API, in the form host:port")
	flagOutput         = flag.String("output", resolvemap.DefaultPath, "path of the resolve map file to write")
	flagInterval       = flag.Duration("interval", time.Minute, "interval to update the resolve map file; if 0, update it once and exit")
)

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	host, port, err := net.SplitHostPort(*flagSabakanAddress)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("invalid sabakan-address: %q (expected host:port)", *flagSabakanAddress)
	}
	if *flagInterval < 0 {
		return fmt.Errorf("invalid interval: %s", *flagInterval)
	}
	sabakan := newSabakanClient(*flagSabakanAddress)

	if *flagInterval == 0 {
		return update(ctx, sabakan, *flagOutput)
	}

	slog.Info("starting", "output", *flagOutput, "interval", flagInterval.String())
	written := false
	for {
		// Keep running on errors such as a temporary failure of sabakan.
		// bmc-reverse-proxy keeps using the last written file in the meantime.
		err := update(ctx, sabakan, *flagOutput)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("failed to update resolve map", "error", err)
		}
		if err == nil {
			written = true
		}

		// bmc-reverse-proxy is not ready until the file is written, so retry soon until the first success.
		interval := *flagInterval
		if !written {
			interval = min(retryInterval, interval)
		}
		select {
		case <-ctx.Done():
			slog.Info("stopped", "reason", context.Cause(ctx).Error())
			return nil
		case <-time.After(interval):
		}
	}
}
