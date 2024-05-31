package hooks

import (
	"context"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-scale-deployment-replica-count,mutating=false,failurePolicy=fail,sideEffects=None,groups="apps",resources=deployments/scale,verbs=update,versions=v1,name=vscaledeploymentreplicacount.kb.io,admissionReviewVersions={v1,v1beta1}

type deploymentReplicaCountScaleValidator struct {
	client  client.Client
	decoder admission.Decoder
}

func NewDeploymentReplicaCountScaleValidator(c client.Client, decoder admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &deploymentReplicaCountScaleValidator{client: c, decoder: decoder}}
}

func (v *deploymentReplicaCountScaleValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	scale := &autoscalingv1.Scale{}
	err := v.decoder.Decode(req, scale)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	deploy := &appsv1.Deployment{}
	err = v.client.Get(ctx, client.ObjectKey{Namespace: scale.GetNamespace(), Name: scale.GetName()}, deploy)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if !isForceReplicaCountTargetDeployment(deploy) {
		return admission.Allowed("not a target")
	}

	if scale.Spec.Replicas != 0 {
		return admission.Denied("replicas must be 0")
	}

	return admission.Allowed("ok")
}
