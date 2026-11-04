package main

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

var _ = Describe("ConfigMap pointer store", func() {
	const (
		namespace = "test-ns"
		name      = "test-pointers"
	)
	var client corev1client.ConfigMapInterface

	BeforeEach(func() {
		client = fake.NewClientset().CoreV1().ConfigMaps(namespace)
	})

	getData := func(ctx context.Context) map[string]string {
		cm, err := client.Get(ctx, name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return cm.Data
	}

	It("creates the ConfigMap and removes pointers of disappeared machines", func(ctx SpecContext) {
		store, err := newConfigMapPointerStore(ctx, client, namespace, name)
		Expect(err).NotTo(HaveOccurred())

		ptr, err := store.get("SERIAL1")
		Expect(err).NotTo(HaveOccurred())
		Expect(ptr).To(Equal(LastPointer{}))

		Expect(store.put("SERIAL1", LastPointer{LastReadId: 3, FirstCreateTime: 100})).To(Succeed())
		Expect(store.put("SERIAL2", LastPointer{LastReadId: 5})).To(Succeed())
		Expect(store.sync(ctx, []Machine{{Serial: "SERIAL1"}, {Serial: "SERIAL2"}})).To(Succeed())
		Expect(getData(ctx)).To(HaveLen(2))

		Expect(store.sync(ctx, []Machine{{Serial: "SERIAL1"}})).To(Succeed())
		data := getData(ctx)
		Expect(data).To(HaveLen(1))
		Expect(data).To(HaveKey("SERIAL1"))

		ptr, err = store.get("SERIAL2")
		Expect(err).NotTo(HaveOccurred())
		Expect(ptr).To(Equal(LastPointer{}))
	})

	It("loads pointers from the existing ConfigMap", func(ctx SpecContext) {
		_, err := client.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Data: map[string]string{
				"SERIAL1": `{"LastReadId":7,"LastError":"","LastHttpStatusCode":0,"FirstCreateTime":200}`,
				"BROKEN":  `{`,
			},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		store, err := newConfigMapPointerStore(ctx, client, namespace, name)
		Expect(err).NotTo(HaveOccurred())

		ptr, err := store.get("SERIAL1")
		Expect(err).NotTo(HaveOccurred())
		Expect(ptr).To(Equal(LastPointer{LastReadId: 7, FirstCreateTime: 200}))

		ptr, err = store.get("BROKEN")
		Expect(err).NotTo(HaveOccurred())
		Expect(ptr).To(Equal(LastPointer{}))
	})

	It("skips serials that are not valid ConfigMap keys", func(ctx SpecContext) {
		store, err := newConfigMapPointerStore(ctx, client, namespace, name)
		Expect(err).NotTo(HaveOccurred())

		Expect(store.put("SERIAL1", LastPointer{LastReadId: 1})).To(Succeed())
		Expect(store.put("INVALID/KEY", LastPointer{LastReadId: 1})).To(Succeed())
		Expect(store.sync(ctx, []Machine{{Serial: "SERIAL1"}, {Serial: "INVALID/KEY"}})).To(Succeed())

		data := getData(ctx)
		Expect(data).To(HaveLen(1))
		Expect(data).To(HaveKey("SERIAL1"))
	})
})
