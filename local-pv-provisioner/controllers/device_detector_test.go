package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockDeleter struct {
	DeleteFunc func(path string) error
}

var _ Deleter = &MockDeleter{}

func (d *MockDeleter) Delete(path string) error {
	return d.DeleteFunc(path)
}

func newDeviceDetectorForTest(nodeName, workingNamespace, defaultPVSpecConfigMap string) *DeviceDetector {
	return &DeviceDetector{
		Client:                 k8sClient,
		reader:                 k8sClient,
		log:                    ctrl.Log.WithName("local-pv-provisioner-test"),
		nodeName:               nodeName,
		interval:               0,
		scheme:                 scheme.Scheme,
		workingNamespace:       workingNamespace,
		defaultPVSpecConfigMap: defaultPVSpecConfigMap,
		deleter: &MockDeleter{
			DeleteFunc: func(path string) error {
				return nil
			},
		},
		availableDevices: prometheus.NewGauge(prometheus.GaugeOpts{}),
		errorDevices:     prometheus.NewGauge(prometheus.GaugeOpts{}),
	}
}

func newPVSpecConfigMap(
	name,
	namespace,
	storageClassName,
	volumeMode,
	fsType,
	deviceDir,
	deviceNameFilter string,
) *corev1.ConfigMap {
	data := map[string]string{
		"storageClassName": storageClassName,
		"volumeMode":       volumeMode,
		"deviceDir":        deviceDir,
		"deviceNameFilter": deviceNameFilter,
	}
	if volumeMode == "Filesystem" {
		data["fsType"] = fsType
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func fetchExistingPVNames(ctx context.Context) ([]string, error) {
	var pvList corev1.PersistentVolumeList
	if err := k8sClient.List(ctx, &pvList); err != nil {
		return nil, err
	}

	// Check that each PV's name is ok
	pvNames := []string{}
	for _, pv := range pvList.Items {
		pvNames = append(pvNames, pv.GetName())
	}

	return pvNames, nil
}

func testDo() {
	pvSpecConfigMapName := "lpp-pv-spec-cm"
	pvSpecConfigMapName2 := "lpp-pv-spec-cm2"
	defaultPVSpecConfigMapName := "lpp-default-pv-spec-cm"
	workingNamespace := "lpp"
	storageClassName := "local-storage"
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: workingNamespace,
		},
	}
	node1 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "192.168.0.1",
			Annotations: map[string]string{
				lppAnnotPVSpecConfigMap: pvSpecConfigMapName,
			},
		},
	}
	node2 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "192.168.0.2",
			Annotations: map[string]string{},
		},
	}

	cleanupResources := func(ctx context.Context, workingNamespace string) error {
		// Remove the PVs
		var pvList corev1.PersistentVolumeList
		if err := k8sClient.List(ctx, &pvList); err != nil {
			return err
		}
		for _, pv := range pvList.Items {
			pv.ObjectMeta.Finalizers = []string{}
			if err := k8sClient.Update(ctx, &pv); err != nil {
				return err
			}
			if err := k8sClient.Delete(ctx, &pv); err != nil {
				return err
			}
		}

		// Remove the ConfigMaps in the working namespace
		if err := k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(workingNamespace)); err != nil {
			return err
		}

		// Annotate Node correctly
		if _, err := ctrl.CreateOrUpdate(ctx, k8sClient, node1, func() error {
			node1.Annotations[lppAnnotPVSpecConfigMap] = pvSpecConfigMapName
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	It("should set up the k8s test environment", func(ctx context.Context) {
		var err error

		_, err = ctrl.CreateOrUpdate(ctx, k8sClient, ns, func() error { return nil })
		Expect(err).NotTo(HaveOccurred())

		_, err = ctrl.CreateOrUpdate(ctx, k8sClient, node1, func() error { return nil })
		Expect(err).NotTo(HaveOccurred())
		_, err = ctrl.CreateOrUpdate(ctx, k8sClient, node2, func() error { return nil })
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable(
		"Checking that PVs are created correctly according to the deviceNameFilter",
		func(ctx context.Context, cmSrc *pvSpec, expectedPVNameSuffixes []string) {
			var err error

			expectedPVNames := []interface{}{}
			for _, suffix := range expectedPVNameSuffixes {
				expectedPVNames = append(expectedPVNames, fmt.Sprintf("local-%s-%s", node1.GetName(), suffix))
			}

			useTestFS(map[string]string{
				"/dev/sda": "dummy",
				"/dev/sdb": "dummy",
				"/dev/sdc": "dummy",
			}, func() {
				cm := newPVSpecConfigMap(
					pvSpecConfigMapName,
					workingNamespace,
					storageClassName,
					cmSrc.volumeMode,
					cmSrc.fsType,
					cmSrc.deviceDir,
					cmSrc.deviceNameFilter,
				)
				_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error { return nil })
				Expect(err).NotTo(HaveOccurred())

				dd := newDeviceDetectorForTest(node1.GetName(), workingNamespace, "")
				dd.do(ctx)

				Eventually(func(g Gomega) {
					pvNames, err := fetchExistingPVNames(ctx)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(pvNames).To(ConsistOf(expectedPVNames...))
				}).Should(Succeed())

				// Clean up the created resources for the successive tests
				Eventually(func() error {
					return cleanupResources(ctx, workingNamespace)
				}).Should(Succeed())
			})
		},
		Entry(
			"Using volumeMode: Block, deviceDir: /dev, deviceNameFilter: .*",
			&pvSpec{volumeMode: "Block", deviceDir: "/dev", deviceNameFilter: ".*"},
			[]string{"sda", "sdb", "sdc"},
		),
		Entry(
			"Using volumeMode: Block, deviceDir: /dev, deviceNameFilter: sd[ab]",
			&pvSpec{volumeMode: "Block", deviceDir: "/dev", deviceNameFilter: "sd[ab]"},
			[]string{"sda", "sdb"},
		),
		Entry(
			"Using volumeMode: Filesystem, fsType: ext4, deviceDir: /dev, deviceNameFilter: .*",
			&pvSpec{volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dev", deviceNameFilter: ".*"},
			[]string{"sda", "sdb", "sdc"},
		),
		Entry(
			"Using volumeMode: Filesystem, deviceDir: /dev, deviceNameFilter: sd[ab]",
			&pvSpec{volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dev", deviceNameFilter: "sd[ab]"},
			[]string{"sda", "sdb"},
		),
	)

	// Perform tests that use only 1 node.
	generateOneNodeTestEntries := func(node, cmName string) []interface{} {
		return []interface{}{
			Entry(
				"Using correct configmap (Block)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Block", "", "/dev", ".*"),
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using correct configmap (Filesystem: ext4)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Filesystem", "ext4", "/dev", ".*"),
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using correct configmap (Filesystem: xfs)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Filesystem", "xfs", "/dev", ".*"),
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using correct configmap (Filesystem: btrfs)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Filesystem", "btrfs", "/dev", ".*"),
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using correct configmap with redundant fsType field",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						"storageClassName": storageClassName,
						"volumeMode":       "Block",
						"fsType":           "ext4", // redundant, should be ignored.
						"deviceDir":        "/dev",
						"deviceNameFilter": ".*",
					},
				},
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using correct configmap with redundant random fields",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						"storageClassName": storageClassName,
						"volumeMode":       "Block",
						"deviceDir":        "/dev",
						"deviceNameFilter": ".*",
						"foo":              "bar", // random redundant field, should be ignored.
					},
				},
				[]string{fmt.Sprintf("local-%s-sda", node)},
			),
			Entry(
				"Using invalid configmap name to get no PVs",
				node,
				newPVSpecConfigMap(cmName+"foo", workingNamespace, storageClassName, "Filesystem", "ext4", "/dev", ".*"),
				[]string{},
			),
			Entry(
				"Using invalid volumeMode",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Foo", "", "/dev", ".*"),
				[]string{},
			),
			Entry(
				"Using invalid fsType",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Filesystem", "ntfs", "/dev", ".*"),
				[]string{},
			),
			Entry(
				"Using invalid deviceDir (no entry)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Block", "", "/foo", ".*"),
				[]string{},
			),
			Entry(
				"Using invalid deviceNameFilter (ill-formed regex)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Block", "", "/dev", "("),
				[]string{},
			),
			Entry(
				"Using invalid deviceNameFilter (no entry)",
				node,
				newPVSpecConfigMap(cmName, workingNamespace, storageClassName, "Block", "", "/dev", "foo"),
				[]string{},
			),
			Entry(
				"Using missing storageClassName",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						// storageClassName is missing
						"volumeMode":       "Block",
						"deviceDir":        "/dev",
						"deviceNameFilter": ".*",
					},
				},
				[]string{},
			),
			Entry(
				"Using missing volumeMode",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						// volumeMode is missing
						"storageClassName": storageClassName,
						"deviceDir":        "/dev",
						"deviceNameFilter": ".*",
					},
				},
				[]string{},
			),
			Entry(
				"Using missing deviceDir",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						// deviceDir is missing
						"storageClassName": storageClassName,
						"volumeMode":       "Block",
						"deviceNameFilter": ".*",
					},
				},
				[]string{},
			),
			Entry(
				"Using missing deviceNameFilter",
				node,
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cmName,
						Namespace: workingNamespace,
					},
					Data: map[string]string{
						// deviceNameFilter is missing
						"storageClassName": storageClassName,
						"volumeMode":       "Block",
						"deviceDir":        "/dev",
					},
				},
				[]string{},
			),
		}
	}

	// oneNodeTestEntries is a array to be used as an argument of DescribeTable.
	oneNodeTestEntries := []interface{}{
		// This function should be applied to every test case in DescribeTable.
		// In the function, we first create the PV Spec ConfigMap. After that,
		// we call do() for either the node1, which has the PV Spec ConfigMap annotation,
		// or the node2, which doesn't have the annotation but is influenced by the default
		// PV Spec ConfigMap. Then, we check if the expected PVs are created correctly.
		// Finally, we clean up the created resources for the successive tests.
		func(ctx context.Context, nodeName string, cm *corev1.ConfigMap, expectedPVNames []string) {
			var err error

			useTestFS(map[string]string{
				"/dev/sda": "dummy",
			}, func() {
				_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error { return nil })
				Expect(err).NotTo(HaveOccurred())

				if nodeName == node1.GetName() {
					dd := newDeviceDetectorForTest(node1.GetName(), workingNamespace, "")
					dd.do(ctx)
				} else {
					dd := newDeviceDetectorForTest(node2.GetName(), workingNamespace, defaultPVSpecConfigMapName)
					dd.do(ctx)
				}

				Eventually(func(g Gomega) {
					pvNames, err := fetchExistingPVNames(ctx)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(pvNames).To(Equal(expectedPVNames))
				}).Should(Succeed())

				// Clean up the created resources for the successive tests
				Eventually(func() error {
					return cleanupResources(ctx, workingNamespace)
				}).Should(Succeed())
			})
		},
	}
	oneNodeTestEntries = append(
		oneNodeTestEntries,
		// appends tests using pv spec configmap via annotations
		generateOneNodeTestEntries(node1.GetName(), pvSpecConfigMapName)...,
	)
	oneNodeTestEntries = append(
		oneNodeTestEntries,
		// appends tests using pv spec configmap via default-pv-spec-configmap.
		generateOneNodeTestEntries(node2.GetName(), defaultPVSpecConfigMapName)...,
	)
	DescribeTable(
		"Checking that PVs are correctly created with various pv spec configmaps attached to nodes via annotations",
		oneNodeTestEntries...,
	)

	It("should handle defaultPVSpecConfigMap correctly with annotations locally attached to the nodes", func(ctx context.Context) {
		useTestFS(map[string]string{
			"/dev/node1":   "dummy",
			"/dev/default": "dummy",
		}, func() {
			var err error

			cmNode1 := newPVSpecConfigMap(pvSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev/", "node1")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cmNode1, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			cmDefault := newPVSpecConfigMap(defaultPVSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev/", "default")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cmDefault, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			dd1 := newDeviceDetectorForTest(node1.GetName(), workingNamespace, defaultPVSpecConfigMapName)
			dd1.do(ctx)
			dd2 := newDeviceDetectorForTest(node2.GetName(), workingNamespace, defaultPVSpecConfigMapName)
			dd2.do(ctx)

			// Check that the PVs are correctly created.
			Eventually(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-node1", "local-192.168.0.2-default"))
			}).Should(Succeed())

			// Clean up the created resources for the successive tests
			Eventually(func() error {
				return cleanupResources(ctx, workingNamespace)
			}).Should(Succeed())
		})
	})

	It("should not create any PVs on a node where existing PVs don't have any annotations, until they are annotated", func(ctx context.Context) {
		useTestFS(map[string]string{
			"/dev/sda": "dummy",
			"/dev/sdb": "dummy",
		}, func() {
			var err error

			cmNode1 := newPVSpecConfigMap(pvSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev", ".*")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cmNode1, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			cmDefault := newPVSpecConfigMap(defaultPVSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev", ".*")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cmDefault, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			// Emulate a PV that is created in lpp <=0.2.x
			block := corev1.PersistentVolumeBlock
			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "local-192.168.0.1-sda",
					Labels: map[string]string{
						"local-pv-provisioner.cybozu.com/node": "192.168.0.1",
					},
				},
				Spec: corev1.PersistentVolumeSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewQuantity(5, resource.BinarySI),
					},
					NodeAffinity: &corev1.VolumeNodeAffinity{
						Required: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      corev1.LabelHostname,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"192.168.0.1"},
								},
							}},
						}},
					},
					PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						Local: &corev1.LocalVolumeSource{Path: "/dev/sda"},
					},
					StorageClassName: storageClassName,
					VolumeMode:       &block,
				},
			}
			err = k8sClient.Create(ctx, pv)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda"))
			}).Should(Succeed())

			dd1 := newDeviceDetectorForTest(node1.GetName(), workingNamespace, defaultPVSpecConfigMapName)
			dd1.do(ctx)

			// Check that the reconciliation stopped.
			Consistently(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda"))
			}, "1s", "2s").Should(Succeed())

			// Annotate the existing PVs
			ctrl.CreateOrUpdate(ctx, k8sClient, pv, func() error {
				pv.ObjectMeta.Annotations = map[string]string{
					lppAnnotStorageClassName: storageClassName,
					lppAnnotVolumeMode:       "Block",
					lppAnnotDeviceDir:        "/dev",
					lppAnnotDeviceNameFilter: ".*",
				}
				return nil
			})

			dd1.do(ctx)

			// Check that the new PV is correctly created.
			Eventually(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda", "local-192.168.0.1-sdb"))
			}, "1s", "2s").Should(Succeed())

			// Clean up the created resources for the successive tests
			Eventually(func() error {
				return cleanupResources(ctx, workingNamespace)
			}).Should(Succeed())
		})
	})

	It("should not create any PVs on a node where pv-spec-configmap is not specified, assuming that no default-pv-spec-configmap is specified", func(ctx context.Context) {
		useTestFS(map[string]string{
			"/dev/node1":              "dummy",
			"/dev/should-not-be-used": "dummy",
		}, func() {
			var err error

			cmNode1 := newPVSpecConfigMap(pvSpecConfigMapName, workingNamespace, storageClassName, "Filesystem", "ext4", "/dev/", "node1")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cmNode1, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			dd1 := newDeviceDetectorForTest(node1.GetName(), workingNamespace, "")
			dd1.do(ctx)
			dd2 := newDeviceDetectorForTest(node2.GetName(), workingNamespace, "")
			dd2.do(ctx)

			// Check that the PVs are correctly created.
			Eventually(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-node1"))
			}).Should(Succeed())

			// Clean up the created resources for the successive tests
			Eventually(func() error {
				return cleanupResources(ctx, workingNamespace)
			}).Should(Succeed())
		})
	})

	DescribeTable(
		"Checking that the device detector reflects the change of pv spec configmap",
		func(ctx context.Context, cmUpdater func(context.Context), finalChecker func(Gomega, context.Context)) {
			useTestFS(map[string]string{
				"/dev/sda":  "dummy",
				"/dev/sdb":  "dummy",
				"/dev/sdc":  "dummy",
				"/dev2/sda": "dummy",
			}, func() {
				var err error

				cm := newPVSpecConfigMap(pvSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev", "sd[ab]")
				_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error { return nil })
				Expect(err).NotTo(HaveOccurred())

				dd := newDeviceDetectorForTest(node1.GetName(), workingNamespace, "")
				dd.do(ctx)

				// Check that the PVs are correctly created.
				Eventually(func(g Gomega) {
					pvNames, err := fetchExistingPVNames(ctx)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda", "local-192.168.0.1-sdb"))
				}).Should(Succeed())

				// Change the pv spec configmap
				cmUpdater(ctx)

				dd.do(ctx)

				// Check that the PVs already created still exist.
				Consistently(func(g Gomega) {
					pvNames, err := fetchExistingPVNames(ctx)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda", "local-192.168.0.1-sdb"))
				}, "2s", "1s").Should(Succeed())

				// Remove PVs
				var pvList corev1.PersistentVolumeList
				err = k8sClient.List(ctx, &pvList)
				Expect(err).NotTo(HaveOccurred())
				for _, pv := range pvList.Items {
					pv.ObjectMeta.Finalizers = []string{}
					err = k8sClient.Update(ctx, &pv)
					Expect(err).NotTo(HaveOccurred())
					err = k8sClient.Delete(ctx, &pv)
					Expect(err).NotTo(HaveOccurred())
				}

				dd.do(ctx)

				// Check that the PVs are correctly created.
				Eventually(finalChecker).WithContext(ctx).Should(Succeed())

				// Clean up the created resources for the successive tests
				Eventually(func() error {
					return cleanupResources(ctx, workingNamespace)
				}).Should(Succeed())
			})
		},
		Entry(
			"Changing storageClassName",
			func(ctx context.Context) {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvSpecConfigMapName,
						Namespace: workingNamespace,
					},
				}
				_, err := ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error {
					cm.Data["storageClassName"] = "local-storage-2"
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			},
			func(g Gomega, ctx context.Context) {
				var pvList corev1.PersistentVolumeList
				err := k8sClient.List(ctx, &pvList)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvList.Items).To(ConsistOf(
					HaveField("Name", "local-192.168.0.1-sda"),
					HaveField("Name", "local-192.168.0.1-sdb")))
				g.Expect(pvList.Items).To(HaveEach(HaveField("Spec.StorageClassName", "local-storage-2")))
			},
		),
		Entry(
			"Changing volumeMode and fsType",
			func(ctx context.Context) {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvSpecConfigMapName,
						Namespace: workingNamespace,
					},
				}
				_, err := ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error {
					cm.Data["volumeMode"] = "Filesystem"
					cm.Data["fsType"] = "ext4"
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			},
			func(g Gomega, ctx context.Context) {
				var pvList corev1.PersistentVolumeList
				err := k8sClient.List(ctx, &pvList)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvList.Items).To(ConsistOf(
					HaveField("Name", "local-192.168.0.1-sda"),
					HaveField("Name", "local-192.168.0.1-sdb")))
				g.Expect(pvList.Items).To(HaveEach(HaveField("Spec.VolumeMode", ptr.To(corev1.PersistentVolumeFilesystem))))
				g.Expect(pvList.Items).To(HaveEach(HaveField("Spec.Local.FSType", ptr.To("ext4"))))
			},
		),
		Entry(
			"Changing deviceDir",
			func(ctx context.Context) {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvSpecConfigMapName,
						Namespace: workingNamespace,
					},
				}
				_, err := ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error {
					cm.Data["deviceDir"] = "/dev2"
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			},
			func(g Gomega, ctx context.Context) {
				var pvList corev1.PersistentVolumeList
				err := k8sClient.List(ctx, &pvList)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvList.Items).To(ConsistOf(HaveField("Name", "local-192.168.0.1-sda")))
				g.Expect(pvList.Items).To(HaveEach(HaveField("Spec.Local.Path", "/dev2/sda")))
			},
		),
		Entry(
			"Changing deviceNameFilter",
			func(ctx context.Context) {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvSpecConfigMapName,
						Namespace: workingNamespace,
					},
				}
				_, err := ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error {
					cm.Data["deviceNameFilter"] = "sdc"
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			},
			func(g Gomega, ctx context.Context) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sdc"))
			},
		),
		Entry(
			"Changing annotation value attached on the Node, instead of updating the configmap itself",
			func(ctx context.Context) {
				var err error

				cm2 := newPVSpecConfigMap(pvSpecConfigMapName2, workingNamespace, storageClassName, "Block", "", "/dev", "sdc")
				_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cm2, func() error { return nil })
				Expect(err).NotTo(HaveOccurred())

				node1 := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: node1.GetName(),
					},
				}
				_, err = ctrl.CreateOrUpdate(ctx, k8sClient, node1, func() error {
					node1.Annotations[lppAnnotPVSpecConfigMap] = pvSpecConfigMapName2
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
			},
			func(g Gomega, ctx context.Context) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sdc"))
			},
		),
	)

	It("should not delete PV when its corresponding device gets deleted", func(ctx context.Context) {
		useTestFS(map[string]string{
			"/dev/sda": "dummy",
		}, func() {
			var err error

			cm := newPVSpecConfigMap(pvSpecConfigMapName, workingNamespace, storageClassName, "Block", "", "/dev", ".*")
			_, err = ctrl.CreateOrUpdate(ctx, k8sClient, cm, func() error { return nil })
			Expect(err).NotTo(HaveOccurred())

			dd := newDeviceDetectorForTest(node1.GetName(), workingNamespace, "")
			dd.do(ctx)

			// Check that the PVs are correctly created.
			Eventually(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda"))
			}).Should(Succeed())

			// Delete /dev/sda here
			err = fs.Remove("/dev/sda")
			Expect(err).NotTo(HaveOccurred())

			dd.do(ctx)

			// Check that the PVs already created still exist.
			Consistently(func(g Gomega) {
				pvNames, err := fetchExistingPVNames(ctx)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pvNames).To(ConsistOf("local-192.168.0.1-sda"))
			}, "2s", "1s").Should(Succeed())

			// Clean up the created resources for the successive tests
			Eventually(func() error {
				return cleanupResources(ctx, workingNamespace)
			}).Should(Succeed())
		})
	})
}

func testHasAnnotsSetByAnotherConfiguration() {
	DescribeTable("Checking if PVs are conflicting", func(pvSpec *pvSpec, alreadyCreatedPVsAnnotations []map[string]string, expected bool) {
		alreadyCreatedPVs := []corev1.PersistentVolume{}
		for _, annotSrc := range alreadyCreatedPVsAnnotations {
			annot := map[string]string{}
			for k, v := range annotSrc {
				annot[k] = v
			}
			alreadyCreatedPVs = append(alreadyCreatedPVs, corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annot,
				},
			})
		}

		got := hasAnnotsSetByAnotherConfiguration(pvSpec, alreadyCreatedPVs)
		Expect(got).To(Equal(expected))
	},
		Entry(
			"Using volumeMode: Block; No created PVs",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{},
			false,
		),
		Entry(
			"Using volumeMode: Filesystem; No created PVs",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{},
			false,
		),

		Entry(
			"Using volumeMode: Block; pvSpec is the same settings as the elements in alreadyCreatedPVs",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			false,
		),
		Entry(
			"Using volumeMode: Filesystem; pvSpec is the same settings as the elements in alreadyCreatedPVs",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			false,
		),

		Entry(
			"Using volumeMode: Block; alreadyCreatedPVs have no annotations",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{{}},
			true,
		),
		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have no annotations",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{{}},
			true,
		),

		Entry(
			"Using volumeMode: Block; alreadyCreatedPVs have a different storageClassName",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage-2", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),
		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have a different storageClassName",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage-2", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),

		Entry(
			"Using volumeMode: Block; alreadyCreatedPVs have a different volumeMode",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),
		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have a different volumeMode",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),

		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have a different fsType",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "xfs", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),

		Entry(
			"Using volumeMode: Block; alreadyCreatedPVs have a different deviceDir",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir2", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),
		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have a different deviceDir",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir2", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),

		Entry(
			"Using volumeMode: Block; alreadyCreatedPVs have a different deviceNameFilter",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", deviceDir: "/dir", deviceNameFilter: ".*2"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Block", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*"},
			},
			true,
		),
		Entry(
			"Using volumeMode: Filesystem; alreadyCreatedPVs have a different deviceNameFilter",
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dir", deviceNameFilter: ".*"},
			[]map[string]string{
				{lppAnnotStorageClassName: "local-storage", lppAnnotVolumeMode: "Filesystem", lppAnnotFSType: "ext4", lppAnnotDeviceDir: "/dir", lppAnnotDeviceNameFilter: ".*2"},
			},
			true,
		),
	)
}

func testParsePVSpecConfigMap() {
	DescribeTable("Parsing pv spec configmap (successful cases)", func(data map[string]string, expectedPVSpec *pvSpec) {
		useTestFS(map[string]string{
			"/dev/sda": "",
		}, func() {
			cm := corev1.ConfigMap{
				Data: data,
			}
			pvSpec, err := parsePVSpecConfigMap(&cm)
			Expect(err).To(BeNil())
			Expect(pvSpec.volumeMode).To(Equal(expectedPVSpec.volumeMode))
			Expect(pvSpec.fsType).To(Equal(expectedPVSpec.fsType))
			Expect(pvSpec.deviceDir).To(Equal(expectedPVSpec.deviceDir))
			Expect(pvSpec.deviceNameFilter).To(Equal(expectedPVSpec.deviceNameFilter))
		})
	},
		Entry(
			"Using volumeMode: Block",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Block", "deviceDir": "/dev", "deviceNameFilter": ".*"},
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", fsType: "", deviceDir: "/dev", deviceNameFilter: ".*"},
		),
		Entry(
			"Using volumeMode: Block, fsType: ignored",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Block", "fsType": "ignored", "deviceDir": "/dev", "deviceNameFilter": ".*"},
			&pvSpec{storageClassName: "local-storage", volumeMode: "Block", fsType: "ignored", deviceDir: "/dev", deviceNameFilter: ".*"},
		),
		Entry(
			"Using volumeMode: Filesystem",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Filesystem", "fsType": "ext4", "deviceDir": "/dev", "deviceNameFilter": ".*"},
			&pvSpec{storageClassName: "local-storage", volumeMode: "Filesystem", fsType: "ext4", deviceDir: "/dev", deviceNameFilter: ".*"},
		),
	)

	DescribeTable("Parsing pv spec configmap (erroneous cases)", func(data map[string]string) {
		useTestFS(map[string]string{
			"/dev/sda": "",
		}, func() {
			cm := corev1.ConfigMap{
				Data: data,
			}
			pvSpec, err := parsePVSpecConfigMap(&cm)
			Expect(pvSpec).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
	},
		Entry(
			"storageClasName is missing",
			map[string]string{"volumeMode": "Block", "deviceDir": "/dev", "deviceNameFilter": ".*"},
		),
		Entry(
			"volumeMode is invalid",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Foo", "deviceDir": "/dev", "deviceNameFilter": ".*"},
		),
		Entry(
			"fsType is invalid",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Filesystem", "fsType": "foo", "deviceDir": "/dev", "deviceNameFilter": ".*"},
		),
		Entry(
			"deviceDir is invalid (not exists)",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Block", "deviceDir": "/this-should-not-exist", "deviceNameFilter": ".*"},
		),
		Entry(
			"deviceDir is invalid (not a directory)",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Block", "deviceDir": "/dev/sda", "deviceNameFilter": ".*"},
		),
		Entry(
			"deviceNameFilter is invalid",
			map[string]string{"storageClassName": "local-storage", "volumeMode": "Block", "deviceDir": "/dev", "deviceNameFilter": "("},
		),
	)
}

func testDeviceDetectorCreatePV() {
	It("should create PV with specified ownerReference", func(ctx context.Context) {
		deviceDir := "dummy"
		deviceNameFilter := ".*"
		storageClassName := "local-storage"

		dd := &DeviceDetector{
			Client:                 k8sClient,
			log:                    ctrl.Log.WithName("local-pv-provisioner-test"),
			nodeName:               "test-node-127.0.0.1",
			interval:               0,
			scheme:                 scheme.Scheme,
			workingNamespace:       "lpp",
			defaultPVSpecConfigMap: "",
		}
		node := new(corev1.Node)
		node.Name = "test-node-127.0.0.1"
		node.UID = "test-uid"

		tests := []struct {
			inputDevice    Device
			expectedPvName string
			volumeMode     string
			fsType         string
		}{
			{
				inputDevice: Device{
					Path:          "/dev/dummy/device",
					CapacityBytes: 512,
				},
				expectedPvName: "local-test-node-127.0.0.1-device",
				volumeMode:     "Block",
			},
			{
				inputDevice: Device{
					Path:          "/dev/crypt-disk/by-path/pci-0000:3c:00.0-sas-exp0x500056b35e77bcff-phy0-lun-0",
					CapacityBytes: 1024,
				},
				expectedPvName: "local-test-node-127.0.0.1-pci-0000-3c-00.0-sas-exp0x500056b35e77bcff-phy0-lun-0",
				volumeMode:     "Block",
			},
			{
				inputDevice: Device{
					Path:          "/dev/dummy/device !\"#$%&'()*+,:;<=>?@[\\]^_`{|}~0123456789.ABCDEFGHIJKLMNOPQRSTUVWXYZ.abcdefghijklmnopqrstuvwxyz",
					CapacityBytes: 2048,
				},
				expectedPvName: "local-test-node-127.0.0.1-device-0123456789.abcdefghijklmnopqrstuvwxyz.abcdefghijklmnopqrstuvwxyz",
				volumeMode:     "Block",
			},
			{
				inputDevice: Device{
					Path:          "/dev/dummy/device-fs",
					CapacityBytes: 512,
				},
				expectedPvName: "local-test-node-127.0.0.1-device-fs",
				volumeMode:     "Filesystem",
				fsType:         "ext4",
			},
		}

		for _, tt := range tests {
			device := tt.inputDevice

			By("creating PV")
			err := dd.createPV(ctx, device, node, &pvSpec{
				storageClassName: storageClassName,
				volumeMode:       tt.volumeMode,
				fsType:           tt.fsType,
				deviceDir:        deviceDir,
				deviceNameFilter: deviceNameFilter,
			})
			Expect(err).NotTo(HaveOccurred())

			By("getting PV")
			pv := new(corev1.PersistentVolume)
			err = dd.Get(ctx, types.NamespacedName{Name: tt.expectedPvName}, pv)
			Expect(err).NotTo(HaveOccurred())

			By("checking volumeMode")
			Expect(string(*pv.Spec.VolumeMode)).To(Equal(tt.volumeMode))

			By("checking PV source")
			localVolume := pv.Spec.PersistentVolumeSource.Local
			Expect(localVolume).NotTo(BeNil())
			Expect(localVolume.Path).To(Equal(device.Path))
			if tt.volumeMode == "Filesystem" {
				Expect(*localVolume.FSType).To(Equal(tt.fsType))
			}

			By("checking labels")
			Expect(pv.ObjectMeta.Labels).To(HaveLen(2))
			Expect(pv.ObjectMeta.Labels).To(HaveKey(lppLegacyLabelKey))
			Expect(pv.ObjectMeta.Labels[lppLegacyLabelKey]).To(Equal(node.Name))
			Expect(pv.ObjectMeta.Labels).To(HaveKey(lppAnnotNode))
			Expect(pv.ObjectMeta.Labels[lppAnnotNode]).To(Equal(node.Name))

			By("checking annotations")
			if tt.volumeMode == "Filesystem" {
				Expect(pv.ObjectMeta.Annotations).To(HaveLen(5))
				Expect(pv.ObjectMeta.Annotations[lppAnnotFSType]).To(Equal(tt.fsType))
			} else {
				Expect(pv.ObjectMeta.Annotations).To(HaveLen(4))
				Expect(pv.ObjectMeta.Annotations[lppAnnotFSType]).To(Equal(""))
			}
			Expect(pv.ObjectMeta.Annotations[lppAnnotStorageClassName]).To(Equal(storageClassName))
			Expect(pv.ObjectMeta.Annotations[lppAnnotVolumeMode]).To(Equal(tt.volumeMode))
			Expect(pv.ObjectMeta.Annotations[lppAnnotDeviceDir]).To(Equal(deviceDir))
			Expect(pv.ObjectMeta.Annotations[lppAnnotDeviceNameFilter]).To(Equal(deviceNameFilter))

			By("checking storageClassName")
			Expect(pv.Spec.StorageClassName).To(Equal(storageClassName))

			By("checking capacity")
			Expect(pv.Spec.Capacity).To(HaveKey(corev1.ResourceStorage))
			capacity := pv.Spec.Capacity[corev1.ResourceStorage]
			Expect(capacity.CmpInt64(device.CapacityBytes)).To(Equal(0))

			By("checking NodeAffinity")
			terms := pv.Spec.NodeAffinity.Required.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(HaveLen(1))
			matchExpression := terms[0].MatchExpressions[0]
			Expect(matchExpression.Key).To(Equal("kubernetes.io/hostname"))
			Expect(matchExpression.Operator).To(Equal(corev1.NodeSelectorOpIn))
			Expect(matchExpression.Values).To(HaveLen(1))
			Expect(matchExpression.Values[0]).To(Equal(node.Name))

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
		err := dd.List(ctx, pvList)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvList.Items).To(HaveLen(len(tests)))
	})
}
