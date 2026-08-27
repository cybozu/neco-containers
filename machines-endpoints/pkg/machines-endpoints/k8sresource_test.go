package main

import (
	"net"
	"testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

func TestUpdateTargetEndpoints(t *testing.T) {
	// setup test client
	testClient := k8sClient{
		k8s:        fake.NewClientset(),
		kubeConfig: &clientcmd.DefaultClientConfig,
	}

	// create EndpointSlices for 3 targets
	targets := []net.IP{
		net.ParseIP("1.1.1.1"),
		net.ParseIP("2.2.2.2"),
		net.ParseIP("3.3.3.3"),
	}
	err := testClient.updateTargetEndpoints(t.Context(), "target-foo", targets, 2, []namedPort{{name: "port-bar", port: 1234}})
	if err != nil {
		t.Fatal(err)
	}

	// check Service
	ns, _, err := testClient.kubeConfig.Namespace()
	if err != nil {
		t.Fatal(err)
	}
	services := testClient.k8s.CoreV1().Services(ns)
	service, err := services.Get(t.Context(), "target-foo", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(service.Spec.Ports) != 1 {
		t.Fatal(`len(service.Spec.Ports) != 1`)
	}
	if service.Spec.Ports[0].Name != "port-bar" {
		t.Error(`service.Spec.Ports[0].Name != "port-bar"`)
	}
	if service.Spec.Ports[0].Port != 1234 {
		t.Error(`service.Spec.Ports[0].Port != 1234`)
	}

	// check EndpointSlice #0
	endpointslices := testClient.k8s.DiscoveryV1().EndpointSlices(ns)
	slice, err := endpointslices.Get(t.Context(), "target-foo-0", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if slice.Labels["endpointslice.kubernetes.io/managed-by"] != "machines-endpoints.cybozu.io" {
		t.Error(`slice.Labels["endpointslice.kubernetes.io/managed-by"] != "machines-endpoints.cybozu.io"`)
	}
	if slice.Labels["kubernetes.io/service-name"] != "target-foo" {
		t.Error(`slice.Labels["kubernetes.io/service-name"] != "target-foo"`)
	}
	if len(slice.Endpoints) != 2 {
		t.Fatal(`len(slice.Endpoints) != 2`)
	}
	if len(slice.Endpoints[0].Addresses) != 1 {
		t.Fatal(`len(slice.Endpoints[0].Addresses) != 1`)
	}
	if slice.Endpoints[0].Addresses[0] != "1.1.1.1" {
		t.Error(`slice.Endpoints[0].Addresses[0] != "1.1.1.1"`)
	}
	if len(slice.Endpoints[1].Addresses) != 1 {
		t.Fatal(`len(slice.Endpoints[1].Addresses) != 1`)
	}
	if slice.Endpoints[1].Addresses[0] != "2.2.2.2" {
		t.Error(`slice.Endpoints[1].Addresses[0] != "2.2.2.2"`)
	}
	if len(slice.Ports) != 1 {
		t.Fatal(`len(slice.Ports) != 1`)
	}
	if slice.Ports[0].Name == nil {
		t.Fatal(`slice.Ports[0].Name == nil`)
	}
	if *slice.Ports[0].Name != "port-bar" {
		t.Error(`*slice.Ports[0].Name != "port-bar"`)
	}
	if slice.Ports[0].Port == nil {
		t.Fatal(`slice.Ports[0].Port == nil`)
	}
	if *slice.Ports[0].Port != 1234 {
		t.Error(`*slice.Ports[0].Port != 1234`)
	}

	// check EndpointSlice #1
	slice, err = endpointslices.Get(t.Context(), "target-foo-1", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(slice.Endpoints) != 1 {
		t.Fatal(`len(slice.Endpoints) != 1`)
	}
	if len(slice.Endpoints[0].Addresses) != 1 {
		t.Fatal(`len(slice.Endpoints[0].Addresses) != 1`)
	}
	if slice.Endpoints[0].Addresses[0] != "3.3.3.3" {
		t.Fatal(`slice.Endpoints[0].Addresses[0] != "3.3.3.3"`)
	}

	// check EndpointSlice #2
	_, err = endpointslices.Get(t.Context(), "target-foo-2", metav1.GetOptions{})
	if !k8serrors.IsNotFound(err) {
		t.Error(`!k8serrors.IsNotFound(err)`)
	}

	// update EndpointSlices for 1 target
	targets = []net.IP{
		net.ParseIP("4.4.4.4"),
	}
	err = testClient.updateTargetEndpoints(t.Context(), "target-foo", targets, 2, []namedPort{{name: "port-bar", port: 1234}})
	if err != nil {
		t.Fatal(err)
	}

	// check EndpointSlice #0
	slice, err = endpointslices.Get(t.Context(), "target-foo-0", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(slice.Endpoints) != 1 {
		t.Fatal(`len(slice.Endpoints) != 1`)
	}
	if len(slice.Endpoints[0].Addresses) != 1 {
		t.Fatal(`len(slice.Endpoints[0].Addresses) != 1`)
	}
	if slice.Endpoints[0].Addresses[0] != "4.4.4.4" {
		t.Fatal(`slice.Endpoints[0].Addresses[0] != "4.4.4.4"`)
	}

	// check EndpointSlice #1
	_, err = endpointslices.Get(t.Context(), "target-foo-1", metav1.GetOptions{})
	if !k8serrors.IsNotFound(err) {
		t.Error(`!k8serrors.IsNotFound(err)`)
	}
}

func TestUpdateTargetEndpointsMultiplePorts(t *testing.T) {
	testClient := k8sClient{
		k8s:        fake.NewClientset(),
		kubeConfig: &clientcmd.DefaultClientConfig,
	}

	targets := []net.IP{net.ParseIP("1.1.1.1")}
	ports := []namedPort{
		{name: "port-bar", port: 2222},
		{name: "port-foo", port: 1111},
	}
	err := testClient.updateTargetEndpoints(t.Context(), "target-multi", targets, 2, ports)
	if err != nil {
		t.Fatal(err)
	}

	ns, _, err := testClient.kubeConfig.Namespace()
	if err != nil {
		t.Fatal(err)
	}

	service, err := testClient.k8s.CoreV1().Services(ns).Get(t.Context(), "target-multi", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(service.Spec.Ports) != 2 {
		t.Fatalf("len(service.Spec.Ports) != 2: %#v", service.Spec.Ports)
	}
	if service.Spec.Ports[0].Name != "port-bar" || service.Spec.Ports[0].Port != 2222 {
		t.Errorf("unexpected service.Spec.Ports[0]: %#v", service.Spec.Ports[0])
	}
	if service.Spec.Ports[1].Name != "port-foo" || service.Spec.Ports[1].Port != 1111 {
		t.Errorf("unexpected service.Spec.Ports[1]: %#v", service.Spec.Ports[1])
	}

	slice, err := testClient.k8s.DiscoveryV1().EndpointSlices(ns).Get(t.Context(), "target-multi-0", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(slice.Ports) != 2 {
		t.Fatalf("len(slice.Ports) != 2: %#v", slice.Ports)
	}
	if *slice.Ports[0].Name != "port-bar" || *slice.Ports[0].Port != 2222 {
		t.Errorf("unexpected slice.Ports[0]: name=%v port=%v", *slice.Ports[0].Name, *slice.Ports[0].Port)
	}
	if *slice.Ports[1].Name != "port-foo" || *slice.Ports[1].Port != 1111 {
		t.Errorf("unexpected slice.Ports[1]: name=%v port=%v", *slice.Ports[1].Name, *slice.Ports[1].Port)
	}

	// update ports on the existing Service
	newPorts := []namedPort{{name: "port-baz", port: 3333}}
	err = testClient.updateTargetEndpoints(t.Context(), "target-multi", targets, 2, newPorts)
	if err != nil {
		t.Fatal(err)
	}
	service, err = testClient.k8s.CoreV1().Services(ns).Get(t.Context(), "target-multi", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(service.Spec.Ports) != 1 || service.Spec.Ports[0].Name != "port-baz" || service.Spec.Ports[0].Port != 3333 {
		t.Errorf("Service ports were not updated: %#v", service.Spec.Ports)
	}
}

func TestDryRun(t *testing.T) {
	testClient := k8sClient{
		k8s:        fake.NewClientset(),
		kubeConfig: &clientcmd.DefaultClientConfig,
		dryRun:     true,
	}

	targets := []net.IP{net.ParseIP("1.1.1.1")}
	ports := []namedPort{{name: "port-bar", port: 1234}}
	if err := testClient.updateTargetEndpoints(t.Context(), "target-foo", targets, 2, ports); err != nil {
		t.Fatal(err)
	}
	if err := testClient.applyConfigMap(t.Context(), "configmap-foo", map[string]string{"key": "value"}); err != nil {
		t.Fatal(err)
	}

	// dry-run must not actually create any resources
	ns, _, err := testClient.kubeConfig.Namespace()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testClient.k8s.CoreV1().Services(ns).Get(t.Context(), "target-foo", metav1.GetOptions{}); !k8serrors.IsNotFound(err) {
		t.Error("dry-run should not create the Service")
	}
	if _, err := testClient.k8s.DiscoveryV1().EndpointSlices(ns).Get(t.Context(), "target-foo-0", metav1.GetOptions{}); !k8serrors.IsNotFound(err) {
		t.Error("dry-run should not create the EndpointSlice")
	}
	if _, err := testClient.k8s.CoreV1().ConfigMaps(ns).Get(t.Context(), "configmap-foo", metav1.GetOptions{}); !k8serrors.IsNotFound(err) {
		t.Error("dry-run should not create the ConfigMap")
	}
}
