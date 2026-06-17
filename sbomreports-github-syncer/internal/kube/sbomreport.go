package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var sbomReportGVR = schema.GroupVersionResource{
	Group:    "aquasecurity.github.io",
	Version:  "v1alpha1",
	Resource: "sbomreports",
}

func ListSbomReports(ctx context.Context, client dynamic.Interface, namespace string, labelSelector string) ([]unstructured.Unstructured, error) {
	listOptions := metav1.ListOptions{LabelSelector: labelSelector}
	resource := client.Resource(sbomReportGVR)

	var (
		list *unstructured.UnstructuredList
		err  error
	)
	if namespace == "" {
		list, err = resource.Namespace(metav1.NamespaceAll).List(ctx, listOptions)
	} else {
		list, err = resource.Namespace(namespace).List(ctx, listOptions)
	}
	if err != nil {
		return nil, fmt.Errorf("list SbomReports: %w", err)
	}
	return list.Items, nil
}
