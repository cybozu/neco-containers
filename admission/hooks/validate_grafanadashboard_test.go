package hooks

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

var _ = Describe("validate GrafanaDashboard webhook with ", func() {
	It("should allow dashboard without any plugins", func() {
		dashboard := `apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: dashboard
  namespace: default
spec:
  instanceSelector:
    matchLabels:
      foo: bar
  json: "{}"
`
		gd := &unstructured.Unstructured{}
		dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, _, err := dec.Decode([]byte(dashboard), nil, gd)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(k8sClient.Create(testCtx, gd)).ShouldNot(HaveOccurred())
	})

	It("should deny dashboard with some plugins", func() {
		dashboardWithPlugins := `apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: dashboard-with-plugins
  namespace: default
spec:
  instanceSelector:
    matchLabels:
      foo: bar
  json: "{}"
  plugins:
    - name: "grafana-piechart-panel"
      version: "1.3.6"
`
		gd := &unstructured.Unstructured{}
		dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, _, err := dec.Decode([]byte(dashboardWithPlugins), nil, gd)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(k8sClient.Create(testCtx, gd)).Should(HaveOccurred())
	})
})
