package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/cybozu/neco-containers/local-pv-provisioner/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}

	// +kubebuilder:scaffold:scheme
}

func run() error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(config.development)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.metricsAddr,
		LeaderElection:     false,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}
	if len(config.nodeName) == 0 {
		err = errors.New("node-name must not by empty")
		setupLog.Error(err, "validation error")
		return err
	}
	if !filepath.IsAbs(config.deviceDir) {
		err = errors.New("device-dir is must be a absolute path")
		setupLog.Error(err, "device-dir is must be a absolute path")
		return err
	}
	info, err := os.Stat(config.deviceDir)
	if err != nil {
		setupLog.Error(err, "unable to get status of divice direcotry", "device-dir", config.deviceDir)
		return err
	}
	if !info.Mode().IsDir() {
		err = errors.New("device-dir is not a direcotry")
		setupLog.Error(err, "divice-dir is not a direcotry")
		return err
	}
	re, err := regexp.Compile(config.deviceNameFilter)
	if err != nil {
		setupLog.Error(err, "unable to compile device filter", "device-name-filter", config.deviceNameFilter)
		return err
	}
	err = mgr.Add(controllers.NewDeviceDetector(mgr.GetClient(), ctrl.Log.WithName("local-pv-provisioner"), config.deviceDir, re, config.nodeName, 10*time.Second, scheme))
	if err != nil {
		setupLog.Error(err, "unable to add device-detector to manager")
		return err
	}

	// pre-cache objects
	if _, err := mgr.GetCache().GetInformer(&corev1.PersistentVolume{}); err != nil {
		return err
	}
	if _, err := mgr.GetCache().GetInformer(&corev1.Node{}); err != nil {
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}
	return nil
}
