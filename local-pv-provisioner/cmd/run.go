package cmd

import (
	"context"
	"errors"

	"github.com/cybozu/neco-containers/local-pv-provisioner/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func run() error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&config.zapOpts)))

	ctx := context.Background()

	if len(config.nodeName) == 0 {
		err := errors.New("node-name must not be empty")
		setupLog.Error(err, "validation error")
		return err
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:         scheme,
		Metrics:        metricsserver.Options{BindAddress: config.metricsAddr},
		LeaderElection: false,
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				config.namespaceName: {},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	deleter := controllers.FillDeleter{
		FillBlockSize: 1024 * 1024,
		FillCount:     100,
	}

	ddLogger := ctrl.Log.WithName("local-pv-provisioner").WithValues("node", config.nodeName)

	dd := controllers.NewDeviceDetector(
		mgr.GetClient(),
		mgr.GetAPIReader(),
		ddLogger,
		config.nodeName,
		config.pollingInterval,
		scheme,
		&deleter,
		config.defaultPVSpecConfigMap,
		config.namespaceName,
	)
	err = mgr.Add(dd)
	if err != nil {
		setupLog.Error(err, "unable to add device-detector to manager")
		return err
	}

	pc := &controllers.PersistentVolumeReconciler{
		Client:   mgr.GetClient(),
		NodeName: config.nodeName,
		Deleter:  &deleter,
	}
	err = pc.SetupWithManager(mgr, config.nodeName)
	if err != nil {
		setupLog.Error(err, "unable to register PersistentVolumeReconciler to mgr")
		return err
	}

	// pre-cache objects
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.PersistentVolume{}); err != nil {
		return err
	}
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.Node{}); err != nil {
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}
	return nil
}
