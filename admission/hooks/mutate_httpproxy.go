package hooks

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	annotationKubernetesIngressClass = "kubernetes.io/ingress.class"
	annotationContourIngressClass    = "projectcontour.io/ingress.class"
	fieldIngressClassName            = "ingressClassName"
	annotationIpPolicy               = annotatePrefix + "ip-policy"
)

// +kubebuilder:webhook:path=/mutate-projectcontour-io-httpproxy,mutating=true,failurePolicy=fail,sideEffects=None,groups=projectcontour.io,resources=httpproxies,verbs=create;update,versions=v1,name=mhttpproxy.kb.io,admissionReviewVersions={v1,v1beta1}

type contourHTTPProxyMutator struct {
	client       client.Client
	decoder      *admission.Decoder
	defaultClass string
	config       *HTTPProxyMutatorConfig
}

// NewContourHTTPProxyMutator creates a webhook handler for Contour HTTPProxy.
func NewContourHTTPProxyMutator(c client.Client, dec *admission.Decoder, defaultClass string, config *HTTPProxyMutatorConfig) http.Handler {
	return &webhook.Admission{Handler: &contourHTTPProxyMutator{c, dec, defaultClass, config}}
}

func getHTTPProxyIngressClassNameField(hp *unstructured.Unstructured) (string, bool, error) {
	return unstructured.NestedString(hp.UnstructuredContent(), "spec", fieldIngressClassName)
}

func setHTTPProxyIngressClassNameField(hp *unstructured.Unstructured, name string) error {
	return unstructured.SetNestedField(hp.UnstructuredContent(), name, "spec", fieldIngressClassName)
}

func (m *contourHTTPProxyMutator) mutateHTTPProxyIngressClassNameField(hp *unstructured.Unstructured) admission.Response {
	if m.defaultClass == "" {
		return admission.Allowed("ok")
	}

	ann := hp.GetAnnotations()

	if _, ok := ann[annotationKubernetesIngressClass]; ok {
		return admission.Allowed("ok")
	}
	if _, ok := ann[annotationContourIngressClass]; ok {
		return admission.Allowed("ok")
	}
	_, ok, err := getHTTPProxyIngressClassNameField(hp)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if ok {
		return admission.Allowed("ok")
	}

	err = setHTTPProxyIngressClassNameField(hp, m.defaultClass)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.Allowed("ok")
}

func getHTTPProxyPolicy(policies []HTTPProxyPolicy, key string) (HTTPProxyPolicy, bool) {
	for _, v := range policies {
		if v.Name == key {
			return v, true
		}
	}
	return HTTPProxyPolicy{}, false
}

func setHTTPProxyPolicyField(hp *unstructured.Unstructured, policy HTTPProxyPolicy) error {
	routes, ok, err := unstructured.NestedSlice(hp.UnstructuredContent(), "spec", "routes")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	for _, route := range routes {
		ipPoliciesOrg, _, err := unstructured.NestedSlice(route.(map[string]interface{}), "ipAllowPolicy")
		if err != nil {
			return err
		}
		ipPolicies := []HTTPProxyIPFilterPolicy{}
		for _, v := range ipPoliciesOrg {
			jsonStr, err := json.Marshal(v)
			if err != nil {
				return err
			}
			ipPolicy := HTTPProxyIPFilterPolicy{}
			if err := json.Unmarshal(jsonStr, &ipPolicy); err != nil {
				return err
			}
			ipPolicies = append(ipPolicies, ipPolicy)
		}
		ipPolicies = append(ipPolicies, policy.IpAllowPolicy...)

		ipPoliciesMap := map[HTTPProxyIPFilterPolicy]struct{}{}
		for _, v := range ipPolicies {
			ipPoliciesMap[v] = struct{}{}
		}
		uniqIpPolicies := []interface{}{}
		for k := range ipPoliciesMap {
			jsonStr, err := json.Marshal(k)
			if err != nil {
				return err
			}
			ipPolicy := map[string]interface{}{}
			if err := json.Unmarshal(jsonStr, &ipPolicy); err != nil {
				return err
			}
			uniqIpPolicies = append(uniqIpPolicies, ipPolicy)
		}
		err = unstructured.SetNestedSlice(route.(map[string]interface{}), uniqIpPolicies, "ipAllowPolicy")
		if err != nil {
			return err
		}
	}
	return unstructured.SetNestedField(hp.UnstructuredContent(), routes, "spec", "routes")
}

func (m *contourHTTPProxyMutator) mutateHTTPProxyPolicy(hp *unstructured.Unstructured) admission.Response {
	ann := hp.GetAnnotations()
	if key, ok := ann[annotationIpPolicy]; ok {
		policy, ok := getHTTPProxyPolicy(m.config.Policies, key)
		if ok {
			err := setHTTPProxyPolicyField(hp, policy)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}
	}
	return admission.Allowed("ok")
}

func (m *contourHTTPProxyMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	hp := &unstructured.Unstructured{}
	err := m.decoder.Decode(req, hp)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if res := m.mutateHTTPProxyIngressClassNameField(hp); !res.Allowed {
		return res
	}

	if res := m.mutateHTTPProxyPolicy(hp); !res.Allowed {
		return res
	}

	marshaled, err := json.Marshal(hp)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
