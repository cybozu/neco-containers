package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"slices"

	discoveryv1 "k8s.io/api/discovery/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	discoveryv1apply "k8s.io/client-go/applyconfigurations/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	labelManagedByValue = "machines-endpoints.cybozu.io"
	fieldManager        = "machines-endpoints"
)

// namedPort is a named port to expose on a target Service and EndpointSlice.
type namedPort struct {
	name string
	port int32
}

type k8sClient struct {
	k8s        kubernetes.Interface
	kubeConfig clientcmd.ClientConfig
	// dryRun makes applyConfigMap/updateTargetEndpoints print the resources
	// that would be applied to stdout, instead of applying them.
	dryRun bool
}

// newKubernetesClient creates a k8sClient. If dryRun is true,
// applyConfigMap/updateTargetEndpoints print the resources that would be
// applied to stdout, instead of applying them.
func newKubernetesClient(dryRun bool) (k8sClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return k8sClient{}, err
	}

	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return k8sClient{}, err
	}

	return k8sClient{k8s: k8sClientSet, kubeConfig: kubeConfig, dryRun: dryRun}, nil
}

// applyConfigMap creates or updates the ConfigMap named name with the given data
// via server-side apply.
func (c k8sClient) applyConfigMap(ctx context.Context, name string, data map[string]string) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	configMapApply := corev1apply.ConfigMap(name, ns).WithData(data)

	if c.dryRun {
		return printResource(configMapApply)
	}

	_, err = c.k8s.CoreV1().ConfigMaps(ns).Apply(ctx, configMapApply, metav1.ApplyOptions{FieldManager: fieldManager, Force: true})
	return err
}

func (c k8sClient) updateTargetEndpoints(ctx context.Context, name string, ips []net.IP, maxEndpoints int, ports []namedPort) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	servicePorts := make([]*corev1apply.ServicePortApplyConfiguration, len(ports))
	endpointPorts := make([]*discoveryv1apply.EndpointPortApplyConfiguration, len(ports))
	for i := range ports {
		p := ports[i]
		servicePorts[i] = corev1apply.ServicePort().WithName(p.name).WithPort(p.port).WithTargetPort(intstr.FromInt(int(p.port)))
		endpointPorts[i] = discoveryv1apply.EndpointPort().WithName(p.name).WithPort(p.port)
	}

	// Create or update the Service
	serviceApply := corev1apply.Service(name, ns).
		WithSpec(corev1apply.ServiceSpec().
			WithPorts(servicePorts...).
			WithClusterIP("None"))
	if c.dryRun {
		if err := printResource(serviceApply); err != nil {
			return err
		}
	} else if _, err := c.k8s.CoreV1().Services(ns).Apply(ctx, serviceApply, metav1.ApplyOptions{FieldManager: fieldManager, Force: true}); err != nil {
		return err
	}

	// Create or update EndpointSlice(s)
	endpointSliceInterface := c.k8s.DiscoveryV1().EndpointSlices(ns)
	sliceIndex := 0
	sliceNames := []string{}
	for chunkedIPs := range slices.Chunk(ips, maxEndpoints) {
		endpoints := make([]*discoveryv1apply.EndpointApplyConfiguration, len(chunkedIPs))
		for i, ip := range chunkedIPs {
			endpoints[i] = discoveryv1apply.Endpoint().WithAddresses(ip.String())
		}

		sliceName := fmt.Sprintf("%s-%d", name, sliceIndex)
		endpointSliceApply := discoveryv1apply.EndpointSlice(sliceName, ns).
			WithLabels(map[string]string{
				discoveryv1.LabelManagedBy:   labelManagedByValue,
				discoveryv1.LabelServiceName: name,
			}).
			WithAddressType(discoveryv1.AddressTypeIPv4).
			WithEndpoints(endpoints...).
			WithPorts(endpointPorts...)

		if c.dryRun {
			if err := printResource(endpointSliceApply); err != nil {
				return err
			}
		} else if _, err := endpointSliceInterface.Apply(ctx, endpointSliceApply, metav1.ApplyOptions{FieldManager: fieldManager, Force: true}); err != nil {
			return err
		}

		sliceIndex++
		sliceNames = append(sliceNames, sliceName)
	}

	if c.dryRun {
		return nil
	}

	// Delete unnecessary EndpointSlice(s)
	sliceList, err := endpointSliceInterface.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", discoveryv1.LabelManagedBy, labelManagedByValue, discoveryv1.LabelServiceName, name),
	})
	if err != nil {
		return err
	}

	for _, slice := range sliceList.Items {
		if slices.Contains(sliceNames, slice.Name) {
			continue
		}
		err := endpointSliceInterface.Delete(ctx, slice.Name, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// printResource prints the JSON representation of an apply configuration,
// instead of applying it.
func printResource(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
