package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func testCiliumIDCollector() {
	It("should report CiliumIdentity count", func() {
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))

			m := make(map[string]int)
			idList := kubectlGetSafe[unstructured.UnstructuredList](g, "ciliumid")
			for _, id := range idList.Items {
				ns := id.GetLabels()["io.kubernetes.pod.namespace"]
				m[ns]++
			}
			for k, v := range m {
				expected := fmt.Sprintf(`neco_cluster_ciliumid_identity_count{namespace="%s"} %d`, k, v)
				g.Expect(output).To(ContainSubstring(expected))
			}
		}).Should(Succeed())
	})

	It("should report CiliumIdentity info", func() {
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))

			var infoLines []string
			for line := range strings.Lines(output) {
				if strings.HasPrefix(line, "neco_cluster_ciliumid_identity_info{") {
					infoLines = append(infoLines, line)
				}
			}

			idList := kubectlGetSafe[unstructured.UnstructuredList](g, "ciliumid")
			g.Expect(idList.Items).NotTo(BeEmpty())
			g.Expect(infoLines).To(HaveLen(len(idList.Items)))

			for _, id := range idList.Items {
				identityLabel := fmt.Sprintf(`identity="%s"`, id.GetName())
				var line string
				for _, l := range infoLines {
					if strings.Contains(l, "{"+identityLabel+",") || strings.Contains(l, ","+identityLabel+",") {
						line = l
						break
					}
				}
				g.Expect(line).NotTo(BeEmpty(), "identity_info not found for %s", id.GetName())
				g.Expect(line).To(ContainSubstring(fmt.Sprintf(`uid="%s"`, id.GetUID())))
				g.Expect(line).To(HaveSuffix("} 1\n"))

				securityLabels, _, err := unstructured.NestedStringMap(id.Object, "security-labels")
				g.Expect(err).NotTo(HaveOccurred())
				if ns, ok := securityLabels["k8s:io.kubernetes.pod.namespace"]; ok {
					g.Expect(line).To(ContainSubstring(fmt.Sprintf(`,namespace="%s"`, ns)))
					g.Expect(line).To(ContainSubstring(fmt.Sprintf(`label_k8s_io_kubernetes_pod_namespace="%s"`, ns)))
				}
			}
		}).Should(Succeed())
	})
}
