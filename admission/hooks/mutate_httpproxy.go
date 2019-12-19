package hooks

import (
	"context"
	"encoding/json"
	"net/http"

	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	annotationKubernetesIngressClass = "kubernetes.io/ingress.class"
	annotationContourIngressClass    = "projectcontour.io/ingress.class"
	annotationIngressClassDefault    = "forest"
)

// +kubebuilder:webhook:verbs=create,path=/mutate-projectcontour-io-httpproxy,mutating=true,failurePolicy=fail,groups=projectcontour.io,resources=httpproxies,versions=v1,name=mhttpproxy.kb.io

type contourHTTPProxyMutator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewContourHTTPProxyMutator creates a webhook handler for Contour HTTPProxy.
func NewContourHTTPProxyMutator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &contourHTTPProxyMutator{c, dec}}
}

func (v *contourHTTPProxyMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	hp := &contourv1.HTTPProxy{}
	err := v.decoder.Decode(req, hp)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Note: An empty class name is not safe if the implementation of Ingress controller considers it as "not specified" and starts serving.
	if hp.Annotations[annotationKubernetesIngressClass] != "" {
		return admission.Allowed("ok")
	}
	if hp.Annotations[annotationContourIngressClass] != "" {
		return admission.Allowed("ok")
	}

	hpPatched := hp.DeepCopy()
	if hpPatched.Annotations == nil {
		hpPatched.Annotations = make(map[string]string)
	}
	hpPatched.Annotations[annotationKubernetesIngressClass] = annotationIngressClassDefault

	marshaled, err := json.Marshal(hpPatched)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
