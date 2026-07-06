package hooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-namespace-deletion,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=namespaces,verbs=delete,versions=v1,name=vnamespacedeletion.kb.io,admissionReviewVersions={v1,v1beta1},timeoutSeconds=10

type namespaceDeletionValidator struct {
	client  client.Client
	decoder admission.Decoder
	mapper  meta.RESTMapper
	config  *NamespaceDeletionValidatorConfig
	logger  logr.Logger
}

// NewNamespaceDeletionValidator creates a webhook handler to validate Namespace DELETE requests.
// It denies deletion if any configured resource in the namespace has the annotation
// `admission.cybozu.com/prevent: delete`.
func NewNamespaceDeletionValidator(c client.Client, dec admission.Decoder, mapper meta.RESTMapper, config *NamespaceDeletionValidatorConfig, logger logr.Logger) (http.Handler, error) {
	v := &namespaceDeletionValidator{
		client:  c,
		decoder: dec,
		mapper:  mapper,
		config:  config,
		logger:  logger,
	}
	return &webhook.Admission{Handler: v}, v.validate()
}

func (v *namespaceDeletionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ns := &corev1.Namespace{}
	if err := v.decoder.DecodeRaw(req.OldObject, ns); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	namespaceName := ns.Name

	for _, r := range v.config.ProtectedResources {
		gvr := schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		}
		gvk, err := v.mapper.KindFor(gvr)
		if err != nil {
			return admission.Errored(
				http.StatusInternalServerError,
				fmt.Errorf("failed to resolve resource %s/%s/%s: %w", r.Group, r.Version, r.Resource, err),
			)
		}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(gvk.GroupVersion().WithKind(gvk.Kind + "List"))

		if err := v.client.List(ctx, list, client.InNamespace(namespaceName)); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		for _, item := range list.Items {
			if item.GetAnnotations()[annPreventKey] == annPreventValueDelete {
				return admission.Denied(fmt.Sprintf("%s %s/%s is protected from deletion", r.Resource, namespaceName, item.GetName()))
			}
		}
	}

	return admission.Allowed("ok")
}

func (v *namespaceDeletionValidator) validate() error {
	if v.config == nil {
		return fmt.Errorf("config must not be nil")
	}

	for _, r := range v.config.ProtectedResources {
		gvr := schema.GroupVersionResource{
			Group:    r.Group,
			Version:  r.Version,
			Resource: r.Resource,
		}

		gvk, err := v.mapper.KindFor(gvr)
		if err != nil {
			return fmt.Errorf("failed to resolve resource %s: %w", gvr.String(), err)
		}

		ssar := &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Group:    r.Group,
					Version:  r.Version,
					Resource: r.Resource,
					Verb:     "list",
				},
			},
		}

		if err := v.client.Create(context.TODO(), ssar); err != nil {
			return fmt.Errorf("failed to create SelfSubjectAccessReview for resource %s: %w", gvr.String(), err)
		}

		if !ssar.Status.Allowed {
			return fmt.Errorf("list permission is not allowed for resource %s", gvr.String())
		}

		v.logger.Info(
			"validated protected resource",
			"group", r.Group,
			"version", r.Version,
			"resource", r.Resource,
			"kind", gvk.Kind,
		)
	}

	return nil
}
