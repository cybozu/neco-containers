package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	calicov3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	annMinimumPolicyOrder = "admission.cybozu.com/min-policy-order"
)

// cnpvlog is for logging in this package.
var cnpvlog = logf.Log.WithName("calico-networkpolicy-validator")

// +kubebuilder:webhook:verbs=create;update,path=/validate-projectcalico-org-networkpolicy,mutating=false,failurePolicy=fail,groups=crd.projectcalico.org,resources=networkpolicies,versions=v1,name=vnetworkpolicy.kb.io
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// CalicoNetworkPolicyValidator is a validating webhook for Calico NetworkPolicy.
// If the order of the policy is equal to or less than the minimum order,
// the validator denies the policy.
//
// The default minimum order is DefaultMinimumOrder.  Each namespace can override
// this with "admission.cybozu.com/min-policy-order" annotation value.
type calicoNetworkPolicyValidator struct {
	client              client.Client
	decoder             *admission.Decoder
	defaultMinimumOrder float64
}

// NewCalicoNetworkPolicyValidator creates a webhook handler for Calico NetworkPolicy.
// The validator denies policies whose order is less than or equal to the given order.
// The default order is minOrder.  This default can be changed per Namespace
// by annotating the namespace with "admission.cybozu.com/min-policy-order".
func NewCalicoNetworkPolicyValidator(c client.Client, dec *admission.Decoder, minOrder float64) http.Handler {
	return &webhook.Admission{Handler: &calicoNetworkPolicyValidator{c, dec, minOrder}}
}

// Handle implements admission.Handler interface.
func (v *calicoNetworkPolicyValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	np := &calicov3.NetworkPolicy{}

	// we cannot use decoder because libcalico-go's api/v3 is badly created.
	// - It does not register resources by `AddToScheme`.
	// - The Group and Version constants defined in the package is not used.
	// Therefore, we need to manually decode the object.
	// err := v.decoder.Decode(req, np)

	if err := json.Unmarshal(req.Object.Raw, np); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	ns := &corev1.Namespace{}
	if err := v.client.Get(ctx, types.NamespacedName{Name: req.Namespace}, ns); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	minOrder := v.defaultMinimumOrder
	if s, ok := ns.Annotations[annMinimumPolicyOrder]; ok {
		min, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// log the error and ignore the annotation
			cnpvlog.Error(err, "non-float value for "+annMinimumPolicyOrder,
				"namespace", req.Namespace)
		} else {
			minOrder = min
		}
	}

	// nil order is handled as positive infinity.
	if np.Spec.Order != nil && *np.Spec.Order <= minOrder {
		return admission.Denied(fmt.Sprintf("order of %s/%s is smaller than required %f < %f",
			req.Namespace, req.Name, *np.Spec.Order, minOrder))
	}
	return admission.Allowed("ok")
}
