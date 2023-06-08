package hooks

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpod.kb.io,admissionReviewVersions={v1,v1beta1}

var (
	ephemeralStorageRequest = resource.MustParse("200Mi")
	ephemeralStorageLimit   = resource.MustParse("1Gi")
)

type podMutator struct {
	client                     client.Client
	decoder                    *admission.Decoder
	ephemeralStoragePermissive bool
}

// NewPodMutator creates a webhook handler for Pod.
func NewPodMutator(c client.Client, dec *admission.Decoder, ephemeralStoragePermissive bool) http.Handler {
	return &webhook.Admission{Handler: &podMutator{c, dec, ephemeralStoragePermissive}}
}

func (m *podMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	po := &corev1.Pod{}
	err := m.decoder.Decode(req, po)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	poPatched := po.DeepCopy()
	for i, co := range po.Spec.Containers {
		if co.Resources.Requests == nil {
			poPatched.Spec.Containers[i].Resources.Requests = corev1.ResourceList{}
		}
		if _, ok := co.Resources.Requests[corev1.ResourceEphemeralStorage]; !ok || !m.ephemeralStoragePermissive {
			poPatched.Spec.Containers[i].Resources.Requests[corev1.ResourceEphemeralStorage] = ephemeralStorageRequest
		}
		if co.Resources.Limits == nil {
			poPatched.Spec.Containers[i].Resources.Limits = corev1.ResourceList{}
		}
		if _, ok := co.Resources.Limits[corev1.ResourceEphemeralStorage]; !ok || !m.ephemeralStoragePermissive {
			poPatched.Spec.Containers[i].Resources.Limits[corev1.ResourceEphemeralStorage] = ephemeralStorageLimit
		}
	}
	for i, co := range po.Spec.InitContainers {
		if co.Resources.Requests == nil {
			poPatched.Spec.InitContainers[i].Resources.Requests = corev1.ResourceList{}
		}
		if _, ok := co.Resources.Requests[corev1.ResourceEphemeralStorage]; !ok || !m.ephemeralStoragePermissive {
			poPatched.Spec.InitContainers[i].Resources.Requests[corev1.ResourceEphemeralStorage] = ephemeralStorageRequest
		}
		if co.Resources.Limits == nil {
			poPatched.Spec.InitContainers[i].Resources.Limits = corev1.ResourceList{}
		}
		if _, ok := co.Resources.Limits[corev1.ResourceEphemeralStorage]; !ok || !m.ephemeralStoragePermissive {
			poPatched.Spec.InitContainers[i].Resources.Limits[corev1.ResourceEphemeralStorage] = ephemeralStorageLimit
		}
	}

	marshaled, err := json.Marshal(poPatched)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
