package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-pod-cpu-request-reduce,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpodcpurequestreduce.kb.io,admissionReviewVersions={v1,v1beta1}

type podCPURequestReducer struct {
	client  client.Client
	decoder admission.Decoder
	logger  logr.Logger
	enabled bool
}

// NewPodCPURequestReducer creates a webhook handler for Pod.
func NewPodCPURequestReducer(c client.Client, dec admission.Decoder, logger logr.Logger, enabled bool) http.Handler {
	return &webhook.Admission{Handler: &podCPURequestReducer{c, dec, logger, enabled}}
}

func reducedRequest(q resource.Quantity) resource.Quantity {
	v := q.ToDec().MilliValue()
	if v == 1 {
		return *resource.NewMilliQuantity(1, resource.DecimalSI)
	} else {
		return *resource.NewMilliQuantity(q.ToDec().MilliValue()/2, resource.DecimalSI)
	}
}

func (m *podCPURequestReducer) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := m.logger.WithValues("namespace", req.Namespace)
	if req.Name != "" {
		log = log.WithValues("podName", req.Name)
	}

	if !m.enabled {
		log.Info("allowed", "reason", "disabled")
		return admission.Allowed("ok")
	}

	po := &corev1.Pod{}
	err := m.decoder.Decode(req, po)
	if err != nil {
		log.Error(err, "unable to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Name == "" {
		log = log.WithValues("podNamePrefix", po.GenerateName)
	}

	for _, owner := range po.GetOwnerReferences() {
		if owner.Kind == "DaemonSet" {
			log.Info("allowed", "reason", "daemonset")
			return admission.Allowed("ok")
		}
	}

	if po.GetLabels()[annotatePrefix+"prevent-cpu-request-reduce"] == "true" {
		log.Info("allowed", "reason", "labeled")
		return admission.Allowed("ok")
	}

	patchedValues := []any{}
	patchedPod := po.DeepCopy()
	for i, co := range po.Spec.InitContainers {
		if q, ok := co.Resources.Requests[corev1.ResourceCPU]; ok {
			reduced := reducedRequest(q)
			patchedValues = append(patchedValues, fmt.Sprintf("initContainer%d", i), fmt.Sprintf("from %s to %s", q.String(), reduced.String()))
			patchedPod.Spec.InitContainers[i].Resources.Requests[corev1.ResourceCPU] = reduced
		}
	}
	for i, co := range po.Spec.Containers {
		if q, ok := co.Resources.Requests[corev1.ResourceCPU]; ok {
			reduced := reducedRequest(q)
			patchedValues = append(patchedValues, fmt.Sprintf("container%d", i), fmt.Sprintf("from %s to %s", q.String(), reduced.String()))
			patchedPod.Spec.Containers[i].Resources.Requests[corev1.ResourceCPU] = reduced
		}
	}

	if len(patchedValues) == 0 {
		log.Info("allowed", "reason", "no cpu request")
		return admission.Allowed("ok")
	}

	marshaled, err := json.Marshal(patchedPod)
	if err != nil {
		log.Error(err, "unable to marshal patched pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Info("patched", patchedValues...)
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
