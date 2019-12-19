package hooks

import (
	"path/filepath"

	calicov3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	// libcalico-go's api/v3 does not implement AddToScheme...
	gv := schema.GroupVersion{Group: "crd.projectcalico.org", Version: "v1"}
	scheme.AddKnownTypes(gv, &calicov3.NetworkPolicy{})
	metav1.AddToGroupVersion(scheme, gv)

	contourv1.AddToScheme(scheme)
}

func run(stopCh <-chan struct{}, cfg *rest.Config, webhookHost string, webhookPort int) error {
	ctrl.SetLogger(zap.Logger(true))

	certDir, err := filepath.Abs("./testdata")
	if err != nil {
		return err
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "localhost:8999",
		LeaderElection:     false,
		Host:               webhookHost,
		Port:               webhookPort,
		CertDir:            certDir,
	})
	if err != nil {
		return err
	}

	dec, _ := admission.NewDecoder(scheme)
	wh := mgr.GetWebhookServer()
	wh.Register("/validate-projectcalico-org-networkpolicy", NewCalicoNetworkPolicyValidator(mgr.GetClient(), dec, 1000))
	wh.Register("/mutate-projectcontour-io-httpproxy", NewContourHTTPProxyMutator(mgr.GetClient(), dec))
	wh.Register("/validate-projectcontour-io-httpproxy", NewContourHTTPProxyValidator(mgr.GetClient(), dec))

	if err := mgr.Start(stopCh); err != nil {
		return err
	}
	return nil
}
