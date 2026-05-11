package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-subnamespace-deletion,mutating=false,failurePolicy=fail,sideEffects=None,groups=accurate.cybozu.com,resources=subnamespaces,verbs=delete,versions=v2,name=vsubnamespacedeletion.kb.io,admissionReviewVersions={v1,v1beta1}

type SubNamespaceDeletionValidator struct {
	client          client.Client
	decoder         admission.Decoder
	discoveryClient discovery.DiscoveryInterface
	dynamicClient   dynamic.Interface
}

func NewSubNamespaceDeletionValidator(
	c client.Client,
	dec admission.Decoder,
	discoveryClient discovery.DiscoveryInterface,
	dynamicClient dynamic.Interface,
) http.Handler {
	return &webhook.Admission{
		Handler: &SubNamespaceDeletionValidator{
			client:          c,
			decoder:         dec,
			discoveryClient: discoveryClient,
			dynamicClient:   dynamicClient,
		},
	}
}

func (v *SubNamespaceDeletionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &unstructured.Unstructured{}
	if err := v.decoder.DecodeRaw(req.OldObject, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	targetNamespace := obj.GetName()
	if targetNamespace == "" {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("SubNamespace name is empty"))
	}

	ns := &corev1.Namespace{}
	if err := v.client.Get(ctx, client.ObjectKey{Name: targetNamespace}, ns); err != nil {
		if apierrors.IsNotFound(err) {
			return admission.Allowed("target namespace does not exist")
		}
		return admission.Errored(http.StatusInternalServerError, err)
	}

	var err error

	blocker, found, err := v.findAnyNamespaceResource(ctx, targetNamespace)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if found {
		return admission.Denied(fmt.Sprintf(
			"cannot delete SubNamespace %q in namespace %q because namespace %q still has resource: %s",
			obj.GetName(),
			obj.GetNamespace(),
			targetNamespace,
			blocker,
		))
	}

	return admission.Allowed("ok")
}
func (v *SubNamespaceDeletionValidator) findAnyNamespaceResource(
	ctx context.Context,
	namespace string,
) (string, bool, error) {
	_, resourceLists, err := v.discoveryClient.ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return "", false, fmt.Errorf("discover namespaced resources: %w", err)
	}

	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return "", false, fmt.Errorf("parse group version %q: %w", resourceList.GroupVersion, err)
		}

		for _, apiResource := range resourceList.APIResources {
			if !shouldCheckAPIResource(apiResource) {
				continue
			}

			gvr := gv.WithResource(apiResource.Name)
			list, err := v.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{
				Limit: 1,
			})
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return "", false, fmt.Errorf("list %s in namespace %q: %w", gvr.String(), namespace, err)
			}

			if len(list.Items) != 0 {
				return formatNamespaceResourceBlocker(gvr, list.Items[0].GetName()), true, nil
			}
		}
	}

	return "", false, nil
}

func shouldCheckAPIResource(apiResource metav1.APIResource) bool {
	if !apiResource.Namespaced {
		return false
	}
	if strings.Contains(apiResource.Name, "/") {
		return false
	}
	return hasAPIResourceVerb(apiResource.Verbs, "list")
}

func hasAPIResourceVerb(verbs []string, target string) bool {
	for _, verb := range verbs {
		if verb == target {
			return true
		}
	}
	return false
}

func formatNamespaceResourceBlocker(gvr schema.GroupVersionResource, name string) string {
	if gvr.Group == "" {
		return fmt.Sprintf("%s/%s", gvr.Resource, name)
	}
	return fmt.Sprintf("%s.%s/%s", gvr.Resource, gvr.Group, name)
}
