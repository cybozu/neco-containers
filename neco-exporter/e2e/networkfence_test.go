package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testNetworkFenceCollector() {
	const fenceName = "test-fence"

	It("should create NetworkFence resource", func() {
		manifest := `
apiVersion: csiaddons.openshift.io/v1alpha1
kind: NetworkFence
metadata:
  name: test-fence
spec:
  driver: rook-ceph.rbd.csi.ceph.com
  fenceState: Fenced
  cidrs:
    - "10.0.0.0/24"
  secret:
    name: test-secret
    namespace: default
`
		Eventually(func(g Gomega) {
			kubectlSafe(g, []byte(manifest), "apply", "-f", "-")
		}).Should(Succeed())
	})

	It("should report networkfence_info metric", func() {
		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))
			expected := fmt.Sprintf(
				`neco_cluster_networkfence_info{driver="rook-ceph.rbd.csi.ceph.com",fence_state="Fenced",name=%q,result=""} 1`,
				fenceName,
			)
			g.Expect(output).To(ContainSubstring(expected))
		}).Should(Succeed())
	})

	It("should report result after status patch", func() {
		patch := `{"status":{"result":"Failed","message":"rpc error: code = DeadlineExceeded desc = context deadline exceeded"}}`
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "patch", "networkfence", fenceName,
				"--subresource=status", "--type=merge", "--patch="+patch)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))
			expected := fmt.Sprintf(
				`neco_cluster_networkfence_info{driver="rook-ceph.rbd.csi.ceph.com",fence_state="Fenced",name=%q,result="Failed"} 1`,
				fenceName,
			)
			g.Expect(output).To(ContainSubstring(expected))
		}).Should(Succeed())
	})

	It("should remove metrics when NetworkFence is deleted", func() {
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "delete", "networkfence", fenceName, "--ignore-not-found")
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := string(scrapeClusterLeader(g))
			g.Expect(output).NotTo(ContainSubstring(fmt.Sprintf(`name=%q`, fenceName)))
		}).Should(Succeed())
	})
}
