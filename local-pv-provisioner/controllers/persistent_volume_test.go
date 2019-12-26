package controllers

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		inputOwnerRef := &metav1.OwnerReference{
			APIVersion: "testVersion",
			Kind:       "testKind",
			Name:       "test-node",
			UID:        "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		}

		By("creating PV")
		err := dd.createPV(context.Background(), device, inputOwnerRef)
		Expect(err).NotTo(HaveOccurred())

		By("confirming PV")
		var pvList corev1.PersistentVolumeList
		err = dd.List(context.Background(), &pvList, client.MatchingLabels{nodeNameLabel: dd.nodeName})
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
		Expect(outputOwnerRef.APIVersion).To(Equal(inputOwnerRef.APIVersion))
		Expect(outputOwnerRef.Kind).To(Equal(inputOwnerRef.Kind))
		Expect(outputOwnerRef.Name).To(Equal(inputOwnerRef.Name))
		Expect(outputOwnerRef.UID).To(Equal(inputOwnerRef.UID))
	})
}
