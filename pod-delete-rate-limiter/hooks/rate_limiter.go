package hooks

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/validate-core-v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=delete,versions=v1,name=vpod.kb.io,admissionReviewVersions=v1

// log is for logging in this package.
var podlog = logf.Log.WithName("PodDeleteRateLimiter")

type podDeleteRateLimiter struct {
	client      client.Client
	decoder     admission.Decoder
	minInterval time.Duration
	username    string
	m           sync.Mutex
	lastDeleted time.Time
}

func NewPodDeleteRateLimiterHttpHandler(client client.Client, decoder admission.Decoder, minInterval time.Duration, username string) http.Handler {
	return &webhook.Admission{Handler: NewPodDeleteRateLimiter(client, decoder, minInterval, username)}
}

func NewPodDeleteRateLimiter(client client.Client, decoder admission.Decoder, minInterval time.Duration, username string) admission.Handler {
	return &podDeleteRateLimiter{
		client:      client,
		decoder:     decoder,
		minInterval: minInterval,
		username:    username,
	}
}

func (v *podDeleteRateLimiter) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Delete {
		podlog.Error(nil, "called with unsupported operation", "operation", req.Operation)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported admission operation"))
	}

	obj := &unstructured.Unstructured{}
	if err := v.decoder.DecodeRaw(req.OldObject, obj); err != nil {
		podlog.Error(err, "could not decode object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	kind := obj.GetKind()
	if kind != "Pod" {
		podlog.Error(nil, "called with unsupported kind", "kind", kind)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported resource kind"))
	}

	name := obj.GetName()
	namespace := obj.GetNamespace()

	if obj.GetDeletionTimestamp() != nil {
		podlog.Info("allowed - deletionTimestamp is already set", "namespace", namespace, "name", name)
		return admission.Allowed("ok")
	}

	if req.UserInfo.Username != v.username {
		podlog.Info("allowed - user is not applied rate limit", "namespace", namespace, "name", name, "username", req.UserInfo.Username)
		return admission.Allowed("ok")
	}

	v.m.Lock()
	defer v.m.Unlock()

	elapsed := time.Since(v.lastDeleted)
	if elapsed < v.minInterval {
		podlog.Info("denied - rate limit reached", "namespace", namespace, "name", name)
		return admission.Denied("rate limited reached")
	}

	dryRun := false
	if req.DryRun != nil {
		dryRun = *req.DryRun
	}
	podlog.Info("allowed - rate limit ok", "namespace", namespace, "name", name, "dry_run", dryRun)
	if !dryRun {
		v.lastDeleted = time.Now()
	}
	return admission.Allowed("ok")
}
