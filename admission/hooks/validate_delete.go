package hooks

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const annotationForDelete = "i-am-sure-to-delete"

// +kubebuilder:webhook:verbs=delete,path=/validate-delete,mutating=false,failurePolicy=fail,groups=apiextensions.k8s.io,resources=customresourcedefinitions,versions=v1;v1beta1,name=vdelete.kb.io

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

	if val, ok := ann[annotationForDelete]; ok && val == name {
		return admission.Allowed("confirmed valid annotation")
	}

	return admission.Denied(`add "i-am-sure-to-delete: <name>" annotation to delete this`)
}
