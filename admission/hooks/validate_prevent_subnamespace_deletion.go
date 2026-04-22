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

// +kubebuilder:webhook:path=/validate-prevent-subnamespace-deletion,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=namespaces,verbs=delete,versions=v1,name=vpreventnsdeletion.kb.io,admissionReviewVersions={v1,v1beta1}

type preventSubNamespaceDeletionValidator struct {
	client  client.Client
	decoder admission.Decoder
	config  *PreventSubNamespaceDeletionValidatorConfig
}

// NewPreventSubNamespaceDeletionValidator creates a webhook handler that denies
// namespace deletion when any of the configured resource types exist in that namespace.
func NewPreventSubNamespaceDeletionValidator(c client.Client, dec admission.Decoder, config *PreventSubNamespaceDeletionValidatorConfig) http.Handler {
	return &webhook.Admission{
		Handler: &preventSubNamespaceDeletionValidator{c, dec, config},
	}
}

func (v *preventSubNamespaceDeletionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	namespaceName := req.Name

	for _, res := range v.config.Resources {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   res.Group,
			Version: res.Version,
			Kind:    res.Kind + "List",
		})
		if err := v.client.List(ctx, list, client.InNamespace(namespaceName), client.Limit(1)); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		if len(list.Items) > 0 {
			return admission.Denied(fmt.Sprintf("namespace %s still has %s resources", namespaceName, res.Kind))
		}
	}

	return admission.Allowed("ok")
}
