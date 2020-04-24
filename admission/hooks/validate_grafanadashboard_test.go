package hooks

import (
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validate GrafanaDashboard webhook with ", func() {
	It("should allow dashboard without any plugins", func() {
		gd := &integreatlyv1alpha1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "valid",
				Namespace: "default",
			},
			Spec: integreatlyv1alpha1.GrafanaDashboardSpec{Json: "{}", Name: "valid.json"},
		}
		Expect(k8sClient.Create(testCtx, gd)).ShouldNot(HaveOccurred())
	})

	It("should deny dashboard with some plugins", func() {
		gd := &integreatlyv1alpha1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "valid",
				Namespace: "default",
			},
			Spec: integreatlyv1alpha1.GrafanaDashboardSpec{
				Json:    "{}",
				Name:    "valid.json",
				Plugins: integreatlyv1alpha1.PluginList{{Name: "foo", Version: "v0.0.1"}}},
		}
		Expect(k8sClient.Create(testCtx, gd)).Should(HaveOccurred())
	})
})
