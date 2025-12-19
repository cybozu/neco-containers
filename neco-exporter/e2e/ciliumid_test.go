package e2e

import (
	"fmt"

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
}
