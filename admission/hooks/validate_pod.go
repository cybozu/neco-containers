package hooks

import (
	"context"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=vpod.kb.io,admissionReviewVersions={v1,v1beta1}
// +kubebuilder:webhook:path=/validate-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods/ephemeralcontainers,verbs=update,versions=v1,name=vpodephemeralcontainer.kb.io,admissionReviewVersions={v1,v1beta1}

type podValidator struct {
	client          client.Client
	decoder         *admission.Decoder
	validPrefixes   []string
	imagePermissive bool
}

// NewPodValidator creates a webhook handler for Pod.
func NewPodValidator(c client.Client, dec *admission.Decoder, validImagePrefixes []string, imagePermissive bool) http.Handler {
	v := &podValidator{
		client:          c,
		decoder:         dec,
		validPrefixes:   validImagePrefixes,
		imagePermissive: imagePermissive,
	}
	return &webhook.Admission{Handler: v}
}

func (v *podValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	po := &corev1.Pod{}
	err := v.decoder.Decode(req, po)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	images := make([]string, 0)
	for _, c := range po.Spec.Containers {
		images = append(images, c.Image)
	}
	for _, c := range po.Spec.InitContainers {
		images = append(images, c.Image)
	}
	for _, c := range po.Spec.EphemeralContainers {
		images = append(images, c.Image)
	}
	var warnings []string
OUTER:
	for _, image := range images {
		for _, prefix := range v.validPrefixes {
			if strings.HasPrefix(image, prefix) {
				continue OUTER
			}
		}

		if v.imagePermissive {
			warnings = append(warnings, "image "+image+" is not trusted")
		} else {
			return admission.Denied("untrustworthy image " + image)
		}
	}

	if len(warnings) > 0 {
		return admission.Allowed("warning").WithWarnings(warnings...)
	}

	return admission.Allowed("ok")
}
