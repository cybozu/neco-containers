package hooks

import (
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	contourv1.AddKnownTypes(scheme)

	// We cannot use AddToScheme() of argoproj/argo-cd
	// because it introduces references to k8s.io/kubernetes, which confuses vendor versions.
}

func run(stopCh <-chan struct{}, cfg *rest.Config, opts *envtest.WebhookInstallOptions) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "localhost:8999",
		LeaderElection:     false,
		Host:               opts.LocalServingHost,
		Port:               opts.LocalServingPort,
		CertDir:            opts.LocalServingCertDir,
	})
	if err != nil {
		return err
	}

	dec, _ := admission.NewDecoder(scheme)
	wh := mgr.GetWebhookServer()
	wh.Register(podMutatingWebhookPath, NewPodMutator(mgr.GetClient(), dec))
	wh.Register(calicoValidateWebhookPath, NewCalicoNetworkPolicyValidator(mgr.GetClient(), dec, 1000))
	wh.Register(contourMutatingWebhookPath, NewContourHTTPProxyMutator(mgr.GetClient(), dec, "secured"))
	wh.Register(contourValidateWebhookPath, NewContourHTTPProxyValidator(mgr.GetClient(), dec))
	wh.Register(argocdValidateWebhookPath, NewArgoCDApplicationValidator(mgr.GetClient(), dec, applicationValidatorConfig))
	wh.Register(grafanaDashboardValidateWebhookPath, NewGrafanaDashboardValidator(mgr.GetClient(), dec))
	wh.Register(deleteValidateWebhookPath, NewDeleteValidator(mgr.GetClient(), dec))
	wh.Register(serviceValidateWebhookPath, NewServiceValidator(mgr.GetClient(), dec))

	if err := mgr.Start(stopCh); err != nil {
		return err
	}
	return nil
}
