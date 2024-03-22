package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"ttypdb/common"

	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const httpServerPort = 8080
const defaultPollIntervalSeconds = 60

var currentNamespace string

var flagSelector = pflag.StringP("selector", "l", "", "Selector to filter Pods")
var flagPollIntervalSeconds = pflag.Int("interval", defaultPollIntervalSeconds, "Polling interval in seconds")

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

	logger.Info("starting ttypdb-sidecar...")

	//initMetrics(logger)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ns, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	currentNamespace = string(ns)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)
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

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", handleReadyz)
	mux.Handle("/metrics", promhttp.Handler())
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", httpServerPort),
		Handler: common.NewProxyHTTPHandler(mux, logger),
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		pollPods(ctx, logger, clientset)
	}()

	wg.Wait()
	logger.Info("termination completed")
	return 0
}

func handleReadyz(http.ResponseWriter, *http.Request) {
	// Nothing to do for now
}
