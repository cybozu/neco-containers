package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	calicov3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	testNS1 = "ns1"
	testNS2 = "ns2"
)

func setupNetworkPolicyResources() {
	// Namespaces
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNS1,
		},
	}
	err := k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())

	ns = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNS2,
			Annotations: map[string]string{
				"admission.cybozu.com/min-policy-order": "2000",
			},
		},
	}
	err = k8sClient.Create(testCtx, ns)
	Expect(err).ShouldNot(HaveOccurred())
}

func testNewNetworkPolicy(ns, name string, order float64) *calicov3.NetworkPolicy {
	np := &calicov3.NetworkPolicy{}
	np.Namespace = ns
	np.Name = name
	if order > 0 {
		np.Spec.Order = &order
	}
	return np
}

var _ = Describe("validate networkpolicy webhook", func() {
	It("should deny policy having order < 1000", func() {
		np := testNewNetworkPolicy(testNS1, "np1", 100)
		err := k8sClient.Create(testCtx, np)
		Expect(err).Should(HaveOccurred())
	})

	It("should deny policy having order == 1000", func() {
		np := testNewNetworkPolicy(testNS1, "np2", 1000)
		err := k8sClient.Create(testCtx, np)
		Expect(err).Should(HaveOccurred())
	})

	It("should allow policy w/o order", func() {
		np := testNewNetworkPolicy(testNS1, "np3", -1)
		err := k8sClient.Create(testCtx, np)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should allow policy having order > 1000", func() {
		np := testNewNetworkPolicy(testNS1, "np4", 2000)
		err := k8sClient.Create(testCtx, np)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should deny updating policy to have order < 1000", func() {
		np := &calicov3.NetworkPolicy{}
		err := k8sClient.Get(testCtx, types.NamespacedName{Namespace: testNS1, Name: "np4"}, np)
		Expect(err).ShouldNot(HaveOccurred())

		var order float64 = 10
		np.Spec.Order = &order
		err = k8sClient.Update(testCtx, np)
		Expect(err).Should(HaveOccurred())
	})

	It("should allow updating policy to have order > 1000", func() {
		np := &calicov3.NetworkPolicy{}
		err := k8sClient.Get(testCtx, types.NamespacedName{Namespace: testNS1, Name: "np4"}, np)
		Expect(err).ShouldNot(HaveOccurred())

		var order float64 = 3000
		np.Spec.Order = &order
		err = k8sClient.Update(testCtx, np)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should respect namespace annotation", func() {
		np := testNewNetworkPolicy(testNS2, "np1", 1500)
		err := k8sClient.Create(testCtx, np)
		Expect(err).Should(HaveOccurred())

		np = testNewNetworkPolicy(testNS2, "np2", 2500)
		err = k8sClient.Create(testCtx, np)
		Expect(err).ShouldNot(HaveOccurred())
	})
})
