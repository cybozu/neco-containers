package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const httpServerPort = 8080

var flagNamespace = pflag.String("namespace", "", "The namespace this program runs in. default: read from serviceaccount file")

var currentNamespace = "default"

func newZapLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}

func main() {
	os.Exit(run())
}

func run() int {
	logger := newZapLogger()
	defer logger.Sync()

	pflag.Parse()
	args := pflag.Args()

	clusterNames := args

	if len(clusterNames) == 0 {
		logger.Error("no MySQLCluster is specified")
		return 1
	}

	if *flagNamespace != "" {
		currentNamespace = *flagNamespace
	} else {
		ns, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			logger.Error("could not read namespace file", zap.Error(err))
			return 1
		}
		currentNamespace = string(ns)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", handleReadyz)
	mux.Handle("/metrics", promhttp.Handler())
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", httpServerPort),
		Handler: NewProxyHTTPHandler(mux, logger),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		go func() {
			<-ctx.Done()
			server.Shutdown(context.Background())
		}()
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			logger.Error("failed to start HTTP server", zap.Error(err))
		}
	}()

	for _, n := range clusterNames {
		n := n
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()
			checker, err := NewChecker(ctx, n, logger)
			if err != nil {
				logger.Error("failed to initialize checker", zap.String("cluster", n), zap.Error(err))
				return
			}
			checker.Run(ctx)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		select {
		case signal := <-signalCh:
			logger.Info("caught signal", zap.String("signal", signal.String()))
		case <-ctx.Done():
		}
	}()

	wg.Wait()
	logger.Info("termination completed")
	return 0
}

func handleReadyz(http.ResponseWriter, *http.Request) {
	// Nothing to do for now
}
