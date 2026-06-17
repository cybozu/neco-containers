package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var namespaceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

func GetNamespaceLabel(
	ctx context.Context,
	client dynamic.Interface,
	namespaceName string,
	labelName string,
) (string, bool, error) {
	if namespaceName == "" {
		return "", false, nil
	}
	if labelName == "" {
		return "", false, nil
	}

	obj, err := client.Resource(namespaceGVR).Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return "", false, fmt.Errorf("get namespace %s: %w", namespaceName, err)
	}

	labels := obj.GetLabels()
	if len(labels) == 0 {
		return "", false, nil
	}

	value, ok := labels[labelName]
	return value, ok, nil
}
