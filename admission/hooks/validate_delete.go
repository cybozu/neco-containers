package hooks

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const annotationForDelete = annotatePrefix + "i-am-sure-to-delete"
const labelDevelopment = "development"

// +kubebuilder:webhook:path=/validate-delete,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=namespaces;resourcequotas,verbs=delete,versions=v1,name=vdelete.kb.io,admissionReviewVersions={v1,v1beta1}

type deleteValidator struct {
	client  client.Client
	decoder admission.Decoder
}

// NewDeleteValidator creates a webhook handler to validate DELETE requests.
func NewDeleteValidator(c client.Client, dec admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &deleteValidator{c, dec}}
}

func (v *deleteValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	// Service accounts authenticate with the username system:serviceaccount:(NAMESPACE):(SERVICEACCOUNT)
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#service-account-tokens
	if req.UserInfo.Username == "system:serviceaccount:accurate:accurate-controller-manager" {
		return admission.Allowed("accurate service account is allowed")
	}

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

	ns := obj.GetNamespace()
	if len(ns) != 0 {
		namespace := &corev1.Namespace{}
		err := v.client.Get(ctx, client.ObjectKey{Name: ns}, namespace)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		if val, ok := namespace.Labels[labelDevelopment]; ok && val == "true" {
			return admission.Allowed("ignore in dev namespaces")
		}
	}

	return admission.Denied(fmt.Sprintf(`add "%si-am-sure-to-delete: %s" annotation to delete this`, annotatePrefix, name))
}
