package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-hnc-x-k8s-io-subnamespaceanchors,mutating=false,failurePolicy=fail,sideEffects=None,groups=hnc.x-k8s.io,resources=subnamespaceanchors,verbs=create;update,versions=v1alpha2,name=vsubnamespaceanchors.kb.io,admissionReviewVersions={v1,v1beta1}

type subnamespaceAnchorValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewSubnamespaceAnchorValidator creates a webhook handler for SubnamespaceAnchor.
func NewSubnamespaceAnchorValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &subnamespaceAnchorValidator{c, dec}}
}

func (v *subnamespaceAnchorValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	subns := &unstructured.Unstructured{}
	if err := v.decoder.Decode(req, subns); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	name := subns.GetName()

	if !strings.HasPrefix(name, "dev-") {
		return admission.Denied(fmt.Sprintf("name should start with dev-, %s is not allowed", name))
	}

	return admission.Allowed("ok")
}
