package hooks

import (
	"context"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const annotationForForceReplicaCount = annotatePrefix + "force-replica-count"

// +kubebuilder:webhook:path=/validate-deployment-replica-count,mutating=false,failurePolicy=fail,sideEffects=None,groups="apps",resources=deployments,verbs=create;update,versions=v1,name=vdeploymentreplicacount.kb.io,admissionReviewVersions={v1,v1beta1}

type deploymentReplicaCountValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewDeploymentReplicaCountValidator returns a webhook handler to validate
// CREATE and UPDATE for Deployment resource.
//
// This webhook denies a resources if the resource has the annotation
// `admission.cybozu.com/force-replica-count: "0"` and its .spec.replicas is not
// zero.
func NewDeploymentReplicaCountValidator(c client.Client, decoder *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &deploymentReplicaCountValidator{client: c, decoder: decoder}}
}

// Handle implements the admission.Handler interface.
func (v *deploymentReplicaCountValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	deploy := &appsv1.Deployment{}
	err := v.decoder.Decode(req, deploy)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if val, ok := deploy.GetAnnotations()[annotationForForceReplicaCount]; !ok || val != "0" {
		return admission.Allowed("not a target")
	}

	if deploy.Spec.Replicas == nil {
		return admission.Denied("replicas cannot be nil")
	}

	if *deploy.Spec.Replicas != 0 {
		return admission.Denied("replicas must be 0")
	}

	return admission.Allowed("ok")
}
