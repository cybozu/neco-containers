package hooks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// We cannot use Application in "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	// because it introduces references to k8s.io/kubernetes, which confuses vendor versions.
	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Kind: "Application", Version: "v1alpha1"})
	err := v.decoder.Decode(req, app)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	repoURL, found, err := unstructured.NestedString(app.UnstructuredContent(), "spec", "source", "repoURL")
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to get spec.resource.repoURL; %w", err))
	}
	if !found {
		return admission.Errored(http.StatusBadRequest, errors.New("spec.source.repoURL not found"))
	}
	project, found, err := unstructured.NestedString(app.UnstructuredContent(), "spec", "project")
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to get spec.project; %w", err))
	}
	if !found {
		return admission.Errored(http.StatusBadRequest, errors.New("spec.project not found"))
	}

	for _, p := range v.findProjects(repoURL) {
		if p == project {
			return admission.Allowed("ok")
		}
	}
	return admission.Denied(fmt.Sprintf("project %q is not allowed for the repository %q", project, repoURL))
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
