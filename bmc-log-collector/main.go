package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type bmcLogWriter interface {
	write(stringJson string, serial string) (err error)
}

func doLogScrapingLoop(config selCollector, logWriter bmcLogWriter) {
	config.httpClient = &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 60 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 60 * time.Second,
			}).DialContext,
		},
	}

	// Set up signal handling
	ctx, cancelCause := context.WithCancelCause(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		cancelCause(fmt.Errorf("%v", sig))
	}()

	// Set interval time
	ticker := time.NewTicker(config.intervalTime)
	defer ticker.Stop()

	// Expose metrics via HTTP
	go metrics("/metrics", ":8080")

	// Scraping loop
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			slog.Error("stopped by", "signal", context.Cause(ctx))
			// Graceful stop when catch SIGTERM
			ticker.Stop()
			wg.Wait()
			return
		case <-ticker.C:
			machinesList, err := readMachineList(config.machinesListDir)
			if err != nil {
				slog.Error("can't read the machine list", "err", err, "path", config.machinesListDir)
				return
			}
			// Start log collector workers by BMCs
			for _, m := range machinesList {
				wg.Go(func() {
					config.collectSystemEventLog(ctx, m, logWriter)
				})
			}
			wg.Wait()
			// Remove pointers of the serials that disappeared in machineList, and persist pointers
			err = config.ptrStore.sync(ctx, machinesList)
			if err != nil {
				slog.Error("failed to sync the pointers", "err", err)
			}
		}
	}
}

// BMC log writer to forward Loki
type logProd struct{}

func (l logProd) write(stringJson string, serial string) error {
	// Use default logger to prevent to mix log messages cross go-routine
	log.Print(stringJson)
	return nil
}

var (
	flgUserFile             *string = pflag.String("bmc-user-json", "/users/neco/bmc-user.json", "User and password of BMC")
	flgUserId               *string = pflag.String("user-id", "support", "User ID of bmc-user-json JSON file")
	flgMachineList          *string = pflag.String("machine-list-json", "/config/machineslist.json", "Target machines list of log scraping")
	flgPointerStore         *string = pflag.String("pointer-store", pointerStoreFile, `Where to store pointers: "file" or "configmap"`)
	flgPointerDir           *string = pflag.String("pointer-dir-path", "/data/pointers", "Data directory of pointer management (for --pointer-store=file)")
	flgPointerConfigMap     *string = pflag.String("pointer-configmap", "bmc-log-collector-pointers", "Name of the ConfigMap in the namespace of the pod to store pointers (for --pointer-store=configmap)")
	flgScrapingIntervalTime *int    = pflag.Int("scraping-interval-time", 300, "Timer(sec) of scraping interval time")
)

const (
	pointerStoreFile      = "file"
	pointerStoreConfigMap = "configmap"

	serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func newPointerStore(ctx context.Context) (pointerStore, error) {
	switch *flgPointerStore {
	case pointerStoreFile:
		return newFilePointerStore(*flgPointerDir), nil
	case pointerStoreConfigMap:
		namespace, err := os.ReadFile(serviceAccountNamespaceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read the namespace: %w", err)
		}
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, err
		}
		ns := strings.TrimSpace(string(namespace))
		return newConfigMapPointerStore(ctx, clientset.CoreV1().ConfigMaps(ns), ns, *flgPointerConfigMap)
	default:
		return nil, fmt.Errorf("unknown pointer store: %q", *flgPointerStore)
	}
}

func main() {
	pflag.Parse()

	// Setup slog
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Read user & password for BMC
	user, err := LoadBMCUserConfig(*flgUserFile)
	if err != nil {
		slog.Error("Can't read the user-list on BMC", "err", err)
		os.Exit(1)
	}

	ptrStore, err := newPointerStore(context.Background())
	if err != nil {
		slog.Error("Can't set up the pointer store", "err", err, "pointer-store", *flgPointerStore)
		os.Exit(1)
	}

	// Setup log scraping loop
	configLc := selCollector{
		machinesListDir: *flgMachineList,
		rfSelPath:       "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
		ptrStore:        ptrStore,
		username:        *flgUserId,
		password:        user.Support.Password.Raw,
		intervalTime:    time.Duration(*flgScrapingIntervalTime) * time.Second,
	}

	// Set BMC log writer
	logWriter := logProd{}
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	slog.Info("bmc-log-collector started", "interval time", *flgScrapingIntervalTime)
	doLogScrapingLoop(configLc, logWriter)
}
