// Command cpu-lending-agent lends the exclusive CPUs of idle MySQL replicas
// to opted-in BestEffort pods and reclaims them on primary promotion.
// See README.md for the design.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/containerd/nri/pkg/stub"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

func main() {
	var (
		nodeName       = flag.String("node-name", os.Getenv("NODE_NAME"), "name of the node this agent runs on (default: NODE_NAME env)")
		kubeconfig     = flag.String("kubeconfig", "", "path to kubeconfig (default: in-cluster config)")
		statePath      = flag.String("cpu-manager-state", "/var/lib/kubelet/cpu_manager_state", "path to the kubelet CPU manager checkpoint")
		sysRoot        = flag.String("sys-root", "/sys", "sysfs mount point for CPU topology")
		lenderSelector = flag.String("lender-selector", "app.kubernetes.io/created-by=moco,app.kubernetes.io/name=mysql", "label selector for lender pods")
		roleLabel      = flag.String("role-label", "moco.cybozu.com/role", "label whose value decides whether a lender is lending")
		lendRoleValue  = flag.String("lend-role-value", "replica", "role label value while which the lender lends its CPUs")
		lenderCtr      = flag.String("lender-container", "mysqld", "container in the lender pod whose exclusive CPUs are lent (empty for all containers)")
		wholeCoresOnly = flag.Bool("lend-whole-cores-only", true, "lend only whole physical cores (all SMT siblings together)")
		resyncPeriod   = flag.Duration("resync-period", 1*time.Minute, "periodic reconcile interval")
		retryInterval  = flag.Duration("retry-interval", 5*time.Second, "retry interval after a failed reconcile")
		listenAddr     = flag.String("listen", ":8080", "address for metrics and health endpoints")
		pluginIdx      = flag.String("nri-plugin-index", "10", "NRI plugin index (ordering among plugins)")
		resourceName   = flag.String("preemptible-millicpu-resource", "cpu-lending.cybozu.io/preemptible-millicpu", "extended resource name: advertised as the node lending capacity, and requesting it marks a pod as a borrower")
		baselineCache  = flag.String("baseline-cache", "/var/lib/cpu-lending-agent/baseline", "file to persist the last known shared pool for fail-closed reclaim across restarts (empty to disable)")
		statusAnno     = flag.String("status-annotation", "cpu-lending.cybozu.io/status", "node annotation key for the human-readable lending summary (empty to disable)")
	)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger, config{
		nodeName:       *nodeName,
		kubeconfig:     *kubeconfig,
		statePath:      *statePath,
		sysRoot:        *sysRoot,
		lenderSelector: *lenderSelector,
		roleLabel:      *roleLabel,
		lendRoleValue:  *lendRoleValue,
		lenderCtr:      *lenderCtr,
		wholeCoresOnly: *wholeCoresOnly,
		resyncPeriod:   *resyncPeriod,
		retryInterval:  *retryInterval,
		listenAddr:     *listenAddr,
		pluginIdx:      *pluginIdx,
		resourceName:   *resourceName,
		baselineCache:  *baselineCache,
		statusAnno:     *statusAnno,
	}); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("agent exited with error", "error", err)
		os.Exit(1)
	}
}

type config struct {
	nodeName       string
	kubeconfig     string
	statePath      string
	sysRoot        string
	lenderSelector string
	roleLabel      string
	lendRoleValue  string
	lenderCtr      string
	resourceName   string
	baselineCache  string
	statusAnno     string
	wholeCoresOnly bool
	resyncPeriod   time.Duration
	retryInterval  time.Duration
	listenAddr     string
	pluginIdx      string
}

func run(logger *slog.Logger, cfg config) error {
	if cfg.nodeName == "" {
		return errors.New("--node-name or NODE_NAME is required")
	}
	lenderSel, err := labels.Parse(cfg.lenderSelector)
	if err != nil {
		return fmt.Errorf("invalid --lender-selector: %w", err)
	}
	if cfg.resourceName == "" {
		return errors.New("--preemptible-millicpu-resource must not be empty")
	}

	restCfg, err := loadRESTConfig(cfg.kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Watch only the pods on this node.
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, cfg.resyncPeriod,
		informers.WithTweakListOptions(func(o *metav1.ListOptions) {
			o.FieldSelector = "spec.nodeName=" + cfg.nodeName
		}))
	podInformer := factory.Core().V1().Pods()

	advertiser := NewNodeAdvertiser(clientset, cfg.nodeName, cfg.resourceName)
	agent := NewAgent(Config{
		NodeName:          cfg.nodeName,
		StatePath:         cfg.statePath,
		LenderSelector:    lenderSel,
		RoleLabel:         cfg.roleLabel,
		LendRoleValue:     cfg.lendRoleValue,
		LenderContainer:   cfg.lenderCtr,
		ResourceName:      cfg.resourceName,
		BaselineCacheFile: cfg.baselineCache,
		StatusAnnotation:  cfg.statusAnno,
		WholeCoresOnly:    cfg.wholeCoresOnly,
		ResyncPeriod:      cfg.resyncPeriod,
		RetryInterval:     cfg.retryInterval,
	}, NewTopology(cfg.sysRoot), podInformer.Lister(), logger)
	agent.SetAdvertiser(advertiser)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	defer broadcaster.Shutdown()
	agent.SetRecorder(broadcaster.NewRecorder(scheme.Scheme,
		corev1.EventSource{Component: "cpu-lending-agent", Host: cfg.nodeName}))

	// Any pod event on this node may change the lending state; reconcile is
	// cheap and level-triggered, so just kick on everything.
	if _, err := podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(any) { agent.Kick() },
		UpdateFunc: func(any, any) { agent.Kick() },
		DeleteFunc: func(any) { agent.Kick() },
	}); err != nil {
		return fmt.Errorf("failed to add pod event handler: %w", err)
	}

	factory.Start(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), podInformer.Informer().HasSynced); !ok {
		return errors.New("failed to sync pod informer cache")
	}
	logger.Info("pod informer synced", "node", cfg.nodeName)

	s, err := stub.New(agent,
		stub.WithPluginName("cpu-lending-agent"),
		stub.WithPluginIdx(cfg.pluginIdx),
		stub.WithOnClose(func() {
			// Lost the containerd connection; exit and let the DaemonSet
			// restart us. Synchronize will re-converge on reconnect.
			logger.Error("NRI connection closed")
			cancel()
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create NRI stub: %w", err)
	}
	agent.SetStub(s)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Addr: cfg.listenAddr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.Run(ctx); err != nil {
			return fmt.Errorf("NRI stub exited: %w", err)
		}
		return nil
	})
	g.Go(func() error { return agent.Run(ctx) })
	g.Go(func() error {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("metrics server exited: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
		s.Stop()
		// Withdraw the advertised capacity so that a permanently removed
		// agent does not leave the scheduler placing borrowers onto a node
		// where nobody manages lending anymore. On a rolling update the new
		// agent re-advertises within its first reconcile.
		if err := advertiser.Ensure(shutdownCtx, 0); err != nil {
			logger.Error("failed to withdraw lending capacity on shutdown", "error", err)
		}
		return ctx.Err()
	})

	agent.Kick()
	logger.Info("cpu-lending-agent started", "node", cfg.nodeName)
	return g.Wait()
}

func loadRESTConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig %s: %w", kubeconfig, err)
		}
		return cfg, nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	return cfg, nil
}
