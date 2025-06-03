package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/tsr-transporter/bmc"
	"github.com/cybozu/neco-containers/tsr-transporter/kintone"
	"github.com/cybozu/neco-containers/tsr-transporter/sabakan"
)

func jobLoopMain(bc *bmc.UserConfig, sa *sabakan.Config, ka *kintone.Config) error {
	// Set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// Set interval time
	ticker := time.NewTicker(time.Duration(cfgIntervalSec) * time.Second)
	defer ticker.Stop()

	// Expose metrics via HTTP
	//go metrics("/metrics", ":8080")

	// Scraping loop
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			slog.Error("stopped by", "signal", context.Cause(ctx))
			// Graceful stop when catch SIGTERM
			ticker.Stop()
			wg.Wait()
			return nil
		case <-ticker.C:
			jobMain(ctx, bc, sa, ka)
		}
	}
}
