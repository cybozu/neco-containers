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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// applog is for logging in this package.
var applog = logf.Log.WithName("application-validator")

// +kubebuilder:webhook:path=/validate-argoproj-io-application,mutating=false,failurePolicy=fail,sideEffects=None,groups=argoproj.io,resources=applications,verbs=create;update,versions=v1alpha1,name=vapplication.kb.io,admissionReviewVersions={v1,v1beta1}

type argocdApplicationValidator struct {
	client               client.Client
	decoder              *admission.Decoder
	config               *ArgoCDApplicationValidatorConfig
	repositoryPermissive bool
}

// NewArgoCDApplicationValidator creates a webhook handler for ArgoCD Application.
func NewArgoCDApplicationValidator(c client.Client, dec *admission.Decoder, config *ArgoCDApplicationValidatorConfig, repositoryPermissive bool) http.Handler {
	v := &argocdApplicationValidator{
		client:               c,
		decoder:              dec,
		config:               config,
		repositoryPermissive: repositoryPermissive,
	}
	return &webhook.Admission{Handler: v}
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

	repoURLs, err := v.extractRepoURLs(app)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	project, found, err := unstructured.NestedString(app.UnstructuredContent(), "spec", "project")
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to get spec.project; %w", err))
	}
	if !found {
		return admission.Errored(http.StatusBadRequest, errors.New("spec.project not found"))
	}

	var warnings []string
OUTER:
	for _, repoURL := range repoURLs {
		for _, p := range v.findProjects(repoURL) {
			if p == project {
				continue OUTER
			}
		}
		reason := fmt.Sprintf("project %q is not allowed for the repository %q", project, repoURL)
		if !v.repositoryPermissive {
			return admission.Denied(reason)
		}
		warnings = append(warnings, reason)
	}
	if len(warnings) != 0 {
		return admission.Allowed("warning").WithWarnings(warnings...)
	}
	return admission.Allowed("ok")
}

func (v *argocdApplicationValidator) extractRepoURLs(app *unstructured.Unstructured) ([]string, error) {
	var repoURLs []string

	repoURL, found, err := unstructured.NestedString(app.UnstructuredContent(), "spec", "source", "repoURL")
	if err != nil {
		return nil, fmt.Errorf("unable to get spec.source.repoURL; %w", err)
	}
	if found {
		repoURLs = append(repoURLs, repoURL)
	}

	sources, found, err := unstructured.NestedSlice(app.UnstructuredContent(), "spec", "sources")
	if err != nil {
		return nil, fmt.Errorf("unable to get spec.sources; %w", err)
	}
	if found {
		for i, source := range sources {
			sourceMap, ok := source.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("spec.sources[%d] should be mapping", i)
			}
			repoURL, found, err := unstructured.NestedString(sourceMap, "repoURL")
			if err != nil {
				return nil, fmt.Errorf("unable to get spec.sources[%d].repoURL; %w", i, err)
			}
			if !found {
				return nil, fmt.Errorf("spec.sources[%d].repoURL not found", i)
			}
			repoURLs = append(repoURLs, repoURL)
		}
	}

	if len(repoURLs) == 0 {
		return nil, errors.New("spec.source.repoURL nor spec.sources not found")
	}

	return repoURLs, nil
}

func (v *argocdApplicationValidator) findProjects(repo string) []string {
	var projects []string
	for _, r := range v.config.Rules {
		if r.Repository != "" && r.RepositoryPrefix != "" {
			// ignore rules specifying both the repository and repositoryPrefix,
			// because it is not what people want to do.
			applog.Info("ignored the rule specifying both the repository and repositoryPrefix", "repository", r.Repository, "repositoryPrefix", r.RepositoryPrefix)
			continue
		}
		if v.ignoreGitSuffix(r.Repository) == v.ignoreGitSuffix(repo) {
			projects = append(projects, r.Projects...)
		}
		if r.RepositoryPrefix != "" && strings.HasPrefix(repo, r.RepositoryPrefix) {
			projects = append(projects, r.Projects...)
		}
	}
	return projects
}

func (v *argocdApplicationValidator) ignoreGitSuffix(s string) string {
	return strings.TrimSuffix(s, ".git")
}
