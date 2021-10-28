package hooks

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-preventdelete,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=persistentvolumeclaims,verbs=delete,versions=v1,name=vpreventdelete.kb.io,admissionReviewVersions={v1,v1beta1}

type preventDeleteValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewPreventDeleteValidator creates a webhook handler to validate DELETE requests
// only for resources annotated with `prevent: delete`.
func NewPreventDeleteValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &preventDeleteValidator{c, dec}}
}

func (v *preventDeleteValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &unstructured.Unstructured{}
	err := v.decoder.DecodeRaw(req.OldObject, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if obj.GetKind() == "PersistentVolumeClaim" &&
		req.UserInfo.Username == "system:serviceaccount:topolvm-system:topolvm-controller" {
		return admission.Allowed("topolvm-controller service account is allowed to delete PVCs")
	}

	if obj.GetAnnotations()[annotatePrefix+"prevent"] == "delete" {
		return admission.Denied(obj.GetName() + " is protected from deletion")
	}

	return admission.Allowed("ok")
}
