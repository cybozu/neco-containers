package hooks

import (
	"context"
	"net/http"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func isAllowed(res *admission.Response) bool {
	return res.Allowed && res.Result.Code == int32(http.StatusOK)
}

func isDenied(res *admission.Response) bool {
	return !res.Allowed && res.Result.Code == int32(http.StatusForbidden)
}

func isBadRequest(res *admission.Response) bool {
	return !res.Allowed && res.Result.Code == int32(http.StatusBadRequest)
}

func TestPodDeleteRateLimiter(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))
	scheme := runtime.NewScheme()
	h := NewPodDeleteRateLimiter(nil, admission.NewDecoder(scheme), time.Second*5, "some-user")

	templateRequest := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Delete,
			OldObject: runtime.RawExtension{
				Raw: []byte(`{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "foo"
	}
}`),
			},
			UserInfo: authenticationv1.UserInfo{
				Username: "some-user",
			},
		},
	}

	for _, op := range []admissionv1.Operation{admissionv1.Create, admissionv1.Update} {
		req := templateRequest
		req.Operation = op
		res := h.Handle(context.Background(), req)
		if !isBadRequest(&res) {
			t.Errorf("handler must return `400 Bad Request` for %s operation\n", op)
		}
	}

	req := templateRequest
	req.OldObject.Raw = []byte("xyz")
	res := h.Handle(context.Background(), req)
	if !isBadRequest(&res) {
		t.Errorf("handler must return `400 Bad Request` for broken objects\n")
	}

	req = templateRequest
	req.OldObject.Raw = []byte(`{
	"apiVersion": "v1",
	"kind": "Node",
	"metadata": {
		"name": "foo"
	}
}`)
	res = h.Handle(context.Background(), req)
	if !isBadRequest(&res) {
		t.Errorf("handler must return `400 Bad Request` for non-Pod objects\n")
	}

	req = templateRequest
	req.DryRun = pointer.Bool(true)
	res = h.Handle(context.Background(), req)
	if !isAllowed(&res) {
		t.Errorf("handler must return `allowed` for the first dry-run request\n")
	}

	req = templateRequest
	res = h.Handle(context.Background(), req)
	if !isAllowed(&res) {
		t.Errorf("handler must return `allowed` for the first non-dry-run request because dry-run requests do not change the handler's internal state\n")
	}
	res = h.Handle(context.Background(), req)
	if !isDenied(&res) {
		t.Errorf("handler must return `denied` for the immediately followed non-dry-run request\n")
	}

	req = templateRequest
	req.DryRun = pointer.Bool(true)
	res = h.Handle(context.Background(), req)
	if !isDenied(&res) {
		t.Errorf("handler must return `denied` for the immediately followed dry-run request too\n")
	}

	req = templateRequest
	req.UserInfo.Username = "another-user"
	res = h.Handle(context.Background(), req)
	if !isAllowed(&res) {
		t.Errorf("handler must return `allowed` for other users\n")
	}

	req = templateRequest
	req.OldObject.Raw = []byte(`{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "foo",
		"deletionTimestamp": "2020-01-01T00:00:00Z"
	}
}`)
	res = h.Handle(context.Background(), req)
	if !isAllowed(&res) {
		t.Errorf("handler must return `allowed` for already-deleted objects\n")
	}

	time.Sleep(time.Second * 5)
	req = templateRequest
	res = h.Handle(context.Background(), req)
	if !isAllowed(&res) {
		t.Errorf("handler must return `allowed` after minimum interval\n")
	}
}
