package manager

import (
	"fmt"
	"log/slog"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/option"
)

var m manager.Manager

func setupLogger() {
	logger := logr.FromSlogHandler(slog.Default().Handler())
	ctrl.SetLogger(logger)
	klog.SetLogger(logger)
}

func IsUsed() bool {
	return m != nil
}

func PeekManager() manager.Manager {
	return m
}

func EnsureManager() (manager.Manager, error) {
	if m != nil {
		return m, nil
	}

	slog.Info("set up controller-runtime")

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("unable to add client-go objects: %w", err)
	}

	metricsAddr := fmt.Sprintf("0.0.0.0:%d", option.ControllerMetricsPort)
	probeAddr := fmt.Sprintf("0.0.0.0:%d", option.ControllerProbePort)

	setupLogger()
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          true,
		LeaderElectionID:        "neco-cluster-exporter",
		LeaderElectionNamespace: option.LeaderElectionNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to start local manager: %w", err)
	}

	m = mgr
	return mgr, nil
}
