package hooks

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-grafana-integreatly-org-grafanadashboard,mutating=false,failurePolicy=fail,sideEffects=None,groups=grafana.integreatly.org,resources=grafanadashboards,verbs=create;update,versions=v1beta1,name=vgrafanadashboard.kb.io,admissionReviewVersions={v1,v1beta1}

type grafanaDashboardValidator struct {
	client  client.Client
	decoder admission.Decoder
}

// NewGrafanaDashboardValidator creates a webhook handler for GrafanaDashboard.
func NewGrafanaDashboardValidator(c client.Client, dec admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &grafanaDashboardValidator{c, dec}}
}

func (v *grafanaDashboardValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	gd := &unstructured.Unstructured{}
	gd.SetGroupVersionKind(schema.GroupVersionKind{Group: "grafana.integreatly.org", Kind: "GrafanaDashboard", Version: "v1beta1"})
	err := v.decoder.Decode(req, gd)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	plugins, _, err := unstructured.NestedSlice(gd.UnstructuredContent(), "spec", "plugins")
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to get spec.plugins; %w", err))
	}

	if len(plugins) > 0 {
		return admission.Denied("spec.plugins must be empty")
	}
	return admission.Allowed("ok")
}
