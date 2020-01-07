package controllers

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
			nodeName:         "test-node-127.0.0.1",
			interval:         0,
			scheme:           scheme.Scheme,
		}
		node := new(corev1.Node)
		node.Name = "test-node-127.0.0.1"
		node.UID = "test-uid"

		tests := []struct {
			inputDevice    Device
			expectedPvName string
		}{
			{
				inputDevice: Device{
					Path:          "/dev/dummy/device",
					CapacityBytes: 512,
				},
				expectedPvName: "local-test-node-127.0.0.1-device",
			},
			{
				inputDevice: Device{
					Path:          "/dev/crypt-disk/by-path/pci-0000:3c:00.0-sas-exp0x500056b35e77bcff-phy0-lun-0",
					CapacityBytes: 1024,
				},
				expectedPvName: "local-test-node-127.0.0.1-pci-0000-3c-00.0-sas-exp0x500056b35e77bcff-phy0-lun-0",
			},
			{
				inputDevice: Device{
					Path:          "/dev/dummy/device !\"#$%&'()*+,:;<=>?@[\\]^_`{|}~0123456789.ABCDEFGHIJKLMNOPQRSTUVWXYZ.abcdefghijklmnopqrstuvwxyz",
					CapacityBytes: 2048,
				},
				expectedPvName: "local-test-node-127.0.0.1-device-0123456789.abcdefghijklmnopqrstuvwxyz.abcdefghijklmnopqrstuvwxyz",
			},
		}

		for _, tt := range tests {
			device := tt.inputDevice

			By("creating PV")
			err := dd.createPV(context.Background(), device, node)
			Expect(err).NotTo(HaveOccurred())

			By("getting PV")
			pv := new(corev1.PersistentVolume)
			err = dd.Get(context.Background(), types.NamespacedName{Name: tt.expectedPvName}, pv)
			Expect(err).NotTo(HaveOccurred())

			By("checking PV source")
			localVolume := pv.Spec.PersistentVolumeSource.Local
			Expect(localVolume).NotTo(BeNil())
			Expect(localVolume.Path).To(Equal(device.Path))

			By("checking storageClassName")
			Expect(pv.Spec.StorageClassName).To(Equal("local-storage"))

			By("checking capacity")
			Expect(pv.Spec.Capacity).To(HaveKey(corev1.ResourceStorage))
			capacity := pv.Spec.Capacity[corev1.ResourceStorage]
			Expect(capacity.CmpInt64(device.CapacityBytes)).To(Equal(0))

			By("checking ownerReferences")
			ownerRefList := pv.GetOwnerReferences()
			Expect(ownerRefList).To(HaveLen(1))

			outputOwnerRef := ownerRefList[0]
			Expect(outputOwnerRef.Kind).To(Equal("Node"))
			Expect(outputOwnerRef.Name).To(Equal(node.Name))
			Expect(outputOwnerRef.UID).To(Equal(node.UID))
		}

		By("checking count of PVs")
		pvList := new(corev1.PersistentVolumeList)
		err := dd.List(context.Background(), pvList)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvList.Items).To(HaveLen(len(tests)))
	})
}
