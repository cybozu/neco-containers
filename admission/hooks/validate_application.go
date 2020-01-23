package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	argocd "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-argoproj-io-application,mutating=false,failurePolicy=fail,groups=argoproj.io,resources=applications,versions=v1alpha1,name=vapplication.kb.io

type argocdApplicationValidator struct {
	client  client.Client
	decoder *admission.Decoder
	config  *ArgoCDApplicationValidatorConfig
}

// NewArgoCDApplicationValidator creates a webhook handler for ArgoCD Application.
func NewArgoCDApplicationValidator(c client.Client, dec *admission.Decoder, config *ArgoCDApplicationValidatorConfig) http.Handler {
	return &webhook.Admission{Handler: &argocdApplicationValidator{c, dec, config}}
}

func (v *argocdApplicationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	app := &argocd.Application{}
	err := v.decoder.Decode(req, app)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Check if spec.project is appropriate for the repository of the application.
	projects := v.findProjects(app.Spec.Source.RepoURL)
	if len(projects) == 0 {
		return admission.Allowed("ok")
	}

	for _, p := range projects {
		if p == app.Spec.Project {
			return admission.Allowed("ok")
		}
	}

	return admission.Denied(fmt.Sprintf("project %q is not allowed for repository %q", app.Spec.Project, app.Spec.Source.RepoURL))
}

func (v *argocdApplicationValidator) findProjects(repo string) []string {
	for _, r := range v.config.Rules {
		if v.ignoreGitSuffix(r.Repository) == v.ignoreGitSuffix(repo) {
			return r.Projects
		}
	}
	return nil
}

func (v *argocdApplicationValidator) ignoreGitSuffix(s string) string {
	return strings.TrimSuffix(s, ".git")
}
