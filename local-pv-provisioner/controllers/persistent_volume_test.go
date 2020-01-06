package controllers

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

func testDeviceDetectorCreatePV() {
	It("should create PV with specified ownerReference", func() {
		re := regexp.MustCompile(".*")

		dd := &DeviceDetector{
			Client:           k8sClient,
			log:              ctrl.Log.WithName("local-pv-provisioner-test"),
			deviceDir:        "dummy",
			deviceNameFilter: re,
			nodeName:         "test-node",
			interval:         0,
			scheme:           scheme.Scheme,
		}
		device := Device{
			Path:          "dummy/device",
			CapacityBytes: 512,
		}
		node := new(corev1.Node)
		node.Name = "test-node"
		node.UID = "test-uid"

		By("creating PV")
		err := dd.createPV(context.Background(), device, node)
		Expect(err).NotTo(HaveOccurred())

		By("confirming PV")
		var pvList corev1.PersistentVolumeList
		err = dd.List(context.Background(), &pvList)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvList.Items).To(HaveLen(1))
		pv := pvList.Items[0]

		localVolume := pv.Spec.PersistentVolumeSource.Local
		Expect(localVolume).NotTo(BeNil())
		Expect(localVolume.Path).To(Equal(device.Path))

		Expect(pv.Spec.Capacity).To(HaveKey(corev1.ResourceStorage))
		capacity := pv.Spec.Capacity[corev1.ResourceStorage]
		Expect(capacity.CmpInt64(device.CapacityBytes)).To(Equal(0))

		ownerRefList := pv.GetOwnerReferences()
		Expect(ownerRefList).To(HaveLen(1))

		outputOwnerRef := ownerRefList[0]
		Expect(outputOwnerRef.Kind).To(Equal("Node"))
		Expect(outputOwnerRef.Name).To(Equal(node.Name))
		Expect(outputOwnerRef.UID).To(Equal(node.UID))
	})
}
