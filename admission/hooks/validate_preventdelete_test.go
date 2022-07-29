package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("prevent DELETE requests", func() {
	It("should allow to delete a PVC w/o the special annotation", func() {
		pvc := &corev1.PersistentVolumeClaim{}
		pvc.Name = "foo1"
		pvc.Namespace = "default"
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.DecimalSI),
		}
		scName := "local-storage"
		pvc.Spec.StorageClassName = &scName
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		err := k8sClient.Create(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should deny to delete a PVC w/ the special annotation", func() {
		pvc := &corev1.PersistentVolumeClaim{}
		pvc.Name = "foo2"
		pvc.Namespace = "default"
		pvc.Annotations = map[string]string{"admission.cybozu.com/prevent": "delete"}
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.DecimalSI),
		}
		scName := "local-storage"
		pvc.Spec.StorageClassName = &scName
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		err := k8sClient.Create(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Delete(testCtx, pvc)
		Expect(err).To(HaveOccurred())
	})

	It("should not deny to delete a PVC by topolvm-controller, even if it has the special annotation", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "topolvm-system",
			},
		}
		err := k8sClient.Create(testCtx, ns)
		Expect(err).NotTo(HaveOccurred())

		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "topolvm-system",
				Name:      "topolvm-controller",
			},
		}
		err = k8sClient.Create(testCtx, sa)
		Expect(err).NotTo(HaveOccurred())

		role := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "topolvm-system:controller",
			},
			Rules: []rbacv1.PolicyRule{
				{
					Verbs:     []string{"delete", "get", "list", "update", "watch"},
					APIGroups: []string{""},
					Resources: []string{"persistentvolumeclaims"},
				},
			},
		}
		err = k8sClient.Create(testCtx, role)
		Expect(err).NotTo(HaveOccurred())

		binding := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "topolvm-system:controller",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "topolvm-controller",
					Namespace: "topolvm-system",
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "topolvm-system:controller",
			},
		}
		err = k8sClient.Create(testCtx, binding)
		Expect(err).NotTo(HaveOccurred())

		pvc := &corev1.PersistentVolumeClaim{}
		pvc.Name = "foo3"
		pvc.Namespace = "default"
		pvc.Annotations = map[string]string{"admission.cybozu.com/prevent": "delete"}
		pvc.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.DecimalSI),
		}
		scName := "local-storage"
		pvc.Spec.StorageClassName = &scName
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}

		err = k8sClient.Create(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())

		cfg := rest.CopyConfig(k8sConfig)
		cfg.Impersonate = rest.ImpersonationConfig{
			UserName: "system:serviceaccount:topolvm-system:topolvm-controller",
			Groups:   []string{"system:authenticated"},
		}
		impersonatedClient, err := client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(impersonatedClient).NotTo(BeNil())

		err = impersonatedClient.Delete(testCtx, pvc)
		Expect(err).NotTo(HaveOccurred())
	})
})
