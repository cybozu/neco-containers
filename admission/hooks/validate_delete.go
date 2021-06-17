package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const annotationForDelete = annotatePrefix + "i-am-sure-to-delete"

// +kubebuilder:webhook:path=/validate-delete,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=namespaces,verbs=delete,versions=v1,name=vdelete.kb.io,admissionReviewVersions={v1,v1beta1}

type deleteValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewDeleteValidator creates a webhook handler to validate DELETE requests.
func NewDeleteValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &deleteValidator{c, dec}}
}

func (v *deleteValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &unstructured.Unstructured{}
	err := v.decoder.DecodeRaw(req.OldObject, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	ann := obj.GetAnnotations()
	name := obj.GetName()

	if obj.GroupVersionKind().Kind == "Namespace" && strings.HasPrefix(name, "dev-") {
		return admission.Allowed("confirmed name started with dev-")
	}

	if val, ok := ann[annotationForDelete]; ok && val == name {
		return admission.Allowed("confirmed valid annotation")
	}

	return admission.Denied(fmt.Sprintf(`add "%si-am-sure-to-delete: %s" annotation to delete this`, annotatePrefix, name))
}
