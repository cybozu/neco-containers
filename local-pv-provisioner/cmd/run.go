package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cybozu/neco-containers/local-pv-provisioner/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}

	// +kubebuilder:scaffold:scheme
}

func run() error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(config.development)))
	log := ctrl.Log.WithName("local-pv-provisioner").WithValues("node", config.nodeName)

	ctx := context.Background()

	if len(config.nodeName) == 0 {
		err := errors.New("node-name must not be empty")
		log.Error(err, "validation error")
		return err
	}
	if !filepath.IsAbs(config.deviceDir) {
		err := errors.New("device-dir must be an absolute path")
		log.Error(err, "device-dir must be an absolute path")
		return err
	}
	info, err := os.Stat(config.deviceDir)
	if err != nil {
		log.Error(err, "unable to get status of device directory", "device-dir", config.deviceDir)
		return err
	}
	if !info.Mode().IsDir() {
		err = errors.New("device-dir is not a directory")
		log.Error(err, "device-dir is not a directory")
		return err
	}
	re, err := regexp.Compile(config.deviceNameFilter)
	if err != nil {
		log.Error(err, "unable to compile device filter", "device-name-filter", config.deviceNameFilter)
		return err
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.metricsAddr,
		LeaderElection:     false,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		return err
	}

	deleter := controllers.FillDeleter{
		FillBlockSize: 1024 * 1024,
		FillCount:     100,
	}

	dd := controllers.NewDeviceDetector(mgr.GetClient(), log,
		config.deviceDir, re, config.nodeName, config.pollingInterval, scheme, &deleter)
	err = mgr.Add(dd)
	if err != nil {
		log.Error(err, "unable to add device-detector to manager")
		return err
	}

	pc := &controllers.PersistentVolumeReconciler{
		Cli:      mgr.GetClient(),
		Log:      log,
		NodeName: config.nodeName,
		Deleter:  &deleter,
	}
	err = pc.SetupWithManager(mgr, config.nodeName)
	if err != nil {
		log.Error(err, "unable to register PersistentVolumeReconciler to mgr")
		return err
	}

	// pre-cache objects
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.PersistentVolume{}); err != nil {
		return err
	}
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.Node{}); err != nil {
		return err
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		return err
	}
	return nil
}
