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

// +kubebuilder:webhook:verbs=create;update,path=/validate-integreatly-org-grafanadashboard,mutating=false,failurePolicy=fail,groups=integreatly.org,resources=grafanadashboards,versions=v1alpha1,name=vgrafanadashboard.kb.io

type grafanaDashboardValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewGrafanaDashboardValidator creates a webhook handler for GrafanaDashboard.
func NewGrafanaDashboardValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &grafanaDashboardValidator{c, dec}}
}

func (v *grafanaDashboardValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	gd := &unstructured.Unstructured{}
	gd.SetGroupVersionKind(schema.GroupVersionKind{Group: "integreatly.org", Kind: "GrafanaDashboard", Version: "v1alpha1"})
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
