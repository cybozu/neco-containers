package cmd

import (
	"github.com/cybozu/neco-containers/admission/hooks"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func run(addr string, port int, conf *hooks.Config) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&config.zapOpts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     config.metricsAddr,
		HealthProbeBindAddress: config.probeAddr,
		LeaderElection:         false,
		Host:                   addr,
		Port:                   port,
		CertDir:                config.certDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	// register webhook handlers
	// admission.NewDecoder never returns non-nil error
	dec := admission.NewDecoder(scheme)
	wh := mgr.GetWebhookServer()
	wh.Register("/mutate-pod", hooks.NewPodMutator(mgr.GetClient(), dec, config.ephemeralStoragePermissive))
	wh.Register("/validate-pod", hooks.NewPodValidator(mgr.GetClient(), dec, config.validImagePrefixes, config.imagePermissive))
	wh.Register("/mutate-projectcontour-io-httpproxy", hooks.NewContourHTTPProxyMutator(mgr.GetClient(), dec, config.httpProxyDefaultClass, &conf.HTTPProxyMutatorConfig))
	wh.Register("/validate-projectcontour-io-httpproxy", hooks.NewContourHTTPProxyValidator(mgr.GetClient(), dec))
	wh.Register("/validate-argoproj-io-application", hooks.NewArgoCDApplicationValidator(mgr.GetClient(), dec, &conf.ArgoCDApplicationValidatorConfig, config.repositoryPermissive))
	wh.Register("/validate-grafana-integreatly-org-grafanadashboard", hooks.NewGrafanaDashboardValidator(mgr.GetClient(), dec))
	wh.Register("/validate-delete", hooks.NewDeleteValidator(mgr.GetClient(), dec))
	wh.Register("/validate-preventdelete", hooks.NewPreventDeleteValidator(mgr.GetClient(), dec))
	wh.Register("/validate-deployment-replica-count", hooks.NewDeploymentReplicaCountValidator(mgr.GetClient(), dec))
	wh.Register("/validate-scale-deployment-replica-count", hooks.NewDeploymentReplicaCountScaleValidator(mgr.GetClient(), dec))

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}
	return nil
}
