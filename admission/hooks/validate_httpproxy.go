package hooks

import (
	"context"
	"fmt"
	"net/http"

	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-projectcontour-io-httpproxy,mutating=false,failurePolicy=fail,groups=projectcontour.io,resources=httpproxies,versions=v1,name=vhttpproxy.kb.io

type contourHTTPProxyValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewContourHTTPProxyValidator creates a webhook handler for Contour HTTPProxy.
func NewContourHTTPProxyValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &contourHTTPProxyValidator{c, dec}}
}

func (v *contourHTTPProxyValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	hp := &contourv1.HTTPProxy{}
	err := v.decoder.Decode(req, hp)
	if err != nil {
		fmt.Println("Errored(http.StatusBadRequest, err)")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if _, ok := hp.Annotations[annotationKubernetesIngressClass]; ok {
		fmt.Println("annotationKubernetesIngressClass: admission.Allowed")
		return admission.Allowed("ok")
	}
	if _, ok := hp.Annotations[annotationContourIngressClass]; ok {
		fmt.Println("annotationContourIngressClass: admission.Allowed")
		return admission.Allowed("ok")
	}

	return admission.Denied(fmt.Sprintf("either %s or %s annotation should be set in %s/%s", annotationKubernetesIngressClass, annotationContourIngressClass, hp.Namespace, hp.Name))
}
