package hooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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
	config  *NamespaceDeletionValidatorConfig
	logger  logr.Logger
}

// NewNamespaceDeletionValidator creates a webhook handler to validate Namespace DELETE requests.
// It denies deletion if any configured resource in the namespace has the annotation
// `admission.cybozu.com/prevent: delete`.
func NewNamespaceDeletionValidator(c client.Client, dec admission.Decoder, config *NamespaceDeletionValidatorConfig, logger logr.Logger) http.Handler {
	v := &namespaceDeletionValidator{
		client: c, 
		decoder: dec, 
		config: config, 
		logger: logger,
	}
	for _, r := range config.ProtectedResources {
		v.logger.Info("protected resource configured", "group", r.Group, "version", r.Version, "kind", r.Kind)
	}
	return &webhook.Admission{Handler: v}
}

func (v *namespaceDeletionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ns := &corev1.Namespace{}
	if err := v.decoder.DecodeRaw(req.OldObject, ns); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	namespaceName := ns.Name

	for _, r := range v.config.ProtectedResources {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   r.Group,
			Version: r.Version,
			Kind:    r.Kind + "List",
		})
		if err := v.client.List(ctx, list, client.InNamespace(namespaceName)); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		for _, item := range list.Items {
			if item.GetAnnotations()[annPreventKey] == annPreventValueDelete {
				return admission.Denied(fmt.Sprintf("%s %s/%s is protected from deletion", r.Kind, namespaceName, item.GetName()))
			}
		}
	}

	return admission.Allowed("ok")
}
