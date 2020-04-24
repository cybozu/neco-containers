package hooks

import (
	"context"
	"net/http"

	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
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
	db := &integreatlyv1alpha1.GrafanaDashboard{}
	err := v.decoder.Decode(req, db)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if len(db.Spec.Plugins) > 0 {
		return admission.Denied("spec.plugins must be empty")
	}

	return admission.Allowed("ok")
}
