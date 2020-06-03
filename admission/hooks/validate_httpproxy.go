package hooks

import (
	"context"
	"fmt"
	"net/http"

	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
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
	if err := v.decoder.Decode(req, hp); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	newAnn := hp.Annotations

	switch req.Operation {
	case admissionv1beta1.Create:
		if newAnn[annotationKubernetesIngressClass] == "" && newAnn[annotationContourIngressClass] == "" {
			return admission.Denied(fmt.Sprintf("either %s or %s annotation should be set", annotationKubernetesIngressClass, annotationContourIngressClass))
		}

	case admissionv1beta1.Update:
		old := &contourv1.HTTPProxy{}
		if err := v.decoder.DecodeRaw(req.OldObject, old); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldAnn := old.Annotations

		if newAnn[annotationKubernetesIngressClass] != oldAnn[annotationKubernetesIngressClass] {
			return admission.Denied("chaning annotation " + annotationKubernetesIngressClass + " is not allowed")
		}
		if newAnn[annotationContourIngressClass] != oldAnn[annotationContourIngressClass] {
			return admission.Denied("chaning annotation " + annotationContourIngressClass + " is not allowed")
		}
	}

	return admission.Allowed("ok")
}
