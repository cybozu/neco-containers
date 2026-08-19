package main

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

func TestUpdateTargetEndpoints(t *testing.T) {
	// setup test client
	testClient := client{
		k8s:        fake.NewClientset(),
		kubeConfig: &clientcmd.DefaultClientConfig,
	}

	// create EndpointSlices for 3 targets
	targets := []net.IP{
		net.ParseIP("1.1.1.1"),
		net.ParseIP("2.2.2.2"),
		net.ParseIP("3.3.3.3"),
	}
	err := testClient.updateTargetEndpoints(t.Context(), targets, 2, "target-foo", "port-bar", 1234)
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
	err = testClient.updateTargetEndpoints(t.Context(), targets, 2, "target-foo", "port-bar", 1234)
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

	// TODO remove transitive code
	// register old-style Endpoints and EndpointSlice
	endpoints := testClient.k8s.CoreV1().Endpoints(ns)
	_, err = endpoints.Create(t.Context(), &corev1.Endpoints{ //nolint:staticcheck
		ObjectMeta: metav1.ObjectMeta{
			Name: "target-foo",
			Labels: map[string]string{
				"endpointslice.kubernetes.io/skip-mirror": "true",
			},
		},
		Subsets: []corev1.EndpointSubset{ //nolint:staticcheck
			{
				Addresses: []corev1.EndpointAddress{
					{IP: "4.4.4.4"},
				},
				Ports: []corev1.EndpointPort{
					{Name: "port-bar", Port: 1234},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = endpointslices.Create(t.Context(), &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: "target-foo",
			Labels: map[string]string{
				"endpointslice.kubernetes.io/managed-by": "machines-endpoints.cybozu.com",
				"kubernetes.io/service-name":             "target-foo",
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{
			{Addresses: []string{"4.4.4.4"}},
		},
		Ports: []discoveryv1.EndpointPort{
			{Name: new("port-bar"), Port: new(int32(1234))},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// call update to delete old-style Endpoints and EndpointSlice
	err = testClient.updateTargetEndpoints(t.Context(), targets, 2, "target-foo", "port-bar", 1234)
	if err != nil {
		t.Fatal(err)
	}

	// check old-style Endpoints and EndpointSlice
	_, err = endpoints.Get(t.Context(), "target-foo", metav1.GetOptions{})
	if !k8serrors.IsNotFound(err) {
		t.Error(`!k8serrors.IsNotFound(err)`)
	}

	_, err = endpointslices.Get(t.Context(), "target-foo", metav1.GetOptions{})
	if !k8serrors.IsNotFound(err) {
		t.Error(`!k8serrors.IsNotFound(err)`)
	}
}

func TestUpdateBMCLogCollectorConfigMap(t *testing.T) {
	var ml []Machine

	var m0 Machine
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.1.1.1")
	m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.2.2.2")
	m0.Spec.BMC.IPv4 = "1.3.3.3"
	m0.Spec.Serial = "ABC123"
	ml = append(ml, m0)

	var m1 Machine
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.1.1.1")
	m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.2.2.2")
	m1.Spec.BMC.IPv4 = "2.3.3.3"
	m1.Spec.Serial = "XYZ123"
	ml = append(ml, m1)

	// expectedJSON is made from ml
	expectedJSON := `[{"serial":"ABC123","bmc_ipv4":"1.3.3.3","node_ipv4":"1.1.1.1"},{"serial":"XYZ123","bmc_ipv4":"2.3.3.3","node_ipv4":"2.1.1.1"}]`
	stringJSON, err := createMachinesList(ml)
	if err != nil {
		t.Fatalf("failed create JSON data %#v", err)
	}
	if !cmp.Equal(stringJSON, expectedJSON) {
		t.Fatalf("Not expected JSON data %v", expectedJSON)
	}
}
