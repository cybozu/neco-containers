package hooks

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-service,mutating=false,failurePolicy=fail,groups="",resources=services,versions=v1,name=vservice.kb.io

type serviceValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewServiceValidator creates a webhook handler to reject Service with the externalIPs field filled.
// Please refer to CVE-2020-8554 https://github.com/kubernetes/kubernetes/issues/97076 for details.
func NewServiceValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &serviceValidator{c, dec}}
}

func (v *serviceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	svc := &corev1.Service{}
	err := v.decoder.Decode(req, svc)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if len(svc.Spec.ExternalIPs) > 0 {
		return admission.Denied(
			fmt.Sprintf(
				"applying Service with externalIPs filled is not allowed: len(externalIPs)=%d",
				len(svc.Spec.ExternalIPs),
			),
		)
	}

	return admission.Allowed("ok")
}
