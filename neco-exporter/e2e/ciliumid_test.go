package e2e

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
					infoLines = append(infoLines, strings.TrimSpace(line))
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
				g.Expect(line).To(HaveSuffix("} 1"))

				securityLabels, _, err := unstructured.NestedStringMap(id.Object, "security-labels")
				g.Expect(err).NotTo(HaveOccurred())
				if ns, ok := securityLabels["k8s:io.kubernetes.pod.namespace"]; ok {
					g.Expect(line).To(ContainSubstring(fmt.Sprintf(`,namespace="%s"`, ns)))
				}
				// the namespace security label is exposed only as "namespace"
				g.Expect(line).NotTo(ContainSubstring("label_k8s_io_kubernetes_pod_namespace="))

				// the format of security_labels is checked against actual Hubble metrics in the test below
				g.Expect(line).To(MatchRegexp(`,security_labels="[^"]+",`))
			}
		}).Should(Succeed())
	})

	It("should follow the lifecycle of CiliumIdentity with new security labels", func() {
		// podName and labelKey should match testdata/ciliumid-pod.yaml
		const (
			podName    = "ciliumid-dynamic-label"
			labelKey   = "identity.neco.cybozu.io/e2e-dynamic"
			metricName = "label_k8s_identity_neco_cybozu_io_e2e_dynamic"
		)
		DeferCleanup(func() {
			_, _, _ = kubectl(nil, "delete", "pod", podName, "--ignore-not-found", "--wait=false")
		})

		By("creating a Pod with a new identity-relevant label")
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "apply", "-f", "testdata/ciliumid-pod.yaml")
		}).Should(Succeed())

		By("checking the label appears in identity_info")
		firstID := expectDynamicLabel(labelKey, metricName, "first")

		By("changing the label value to allocate another identity")
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "label", "pod", podName, "--overwrite", labelKey+"=second")
		}).Should(Succeed())

		By("checking the new label value appears in identity_info")
		secondID := expectDynamicLabel(labelKey, metricName, "second")

		By("deleting the Pod to make the identities unused")
		Eventually(func(g Gomega) {
			kubectlSafe(g, nil, "delete", "pod", podName, "--ignore-not-found")
		}).Should(Succeed())

		By("checking identity_info is removed after Cilium garbage-collects the identities")
		for _, id := range []string{firstID, secondID} {
			Eventually(func(g Gomega) {
				_, stderr, err := kubectl(nil, "get", "ciliumid", id)
				g.Expect(err).To(HaveOccurred())
				g.Expect(string(stderr)).To(ContainSubstring("NotFound"))

				output := string(scrapeClusterLeader(g))
				g.Expect(output).NotTo(ContainSubstring(fmt.Sprintf(`neco_cluster_ciliumid_identity_info{identity="%s",`, id)))
			}).Should(Succeed())
		}
	})

	It("should report security_labels compatible with Hubble metrics", func() {
		Eventually(func(g Gomega) {
			pilotID := getEndpointIdentity(g, "default", "app=pilot")
			exporterID := getEndpointIdentity(g, "neco-exporter", "app.kubernetes.io/name=neco-cluster-exporter")

			// scraping the leader generates traffic from pilot to the exporter.
			output := string(scrapeClusterLeader(g))
			pilotLabels := findSecurityLabels(g, output, pilotID)
			exporterLabels := findSecurityLabels(g, output, exporterID)

			source := fmt.Sprintf(`source="%s"`, pilotLabels)
			destination := fmt.Sprintf(`destination="%s"`, exporterLabels)
			var found bool
			for line := range strings.Lines(string(scrapeHubble(g))) {
				if strings.HasPrefix(line, "hubble_flows_processed_total{") &&
					strings.Contains(line, source) && strings.Contains(line, destination) {
					found = true
					break
				}
			}
			g.Expect(found).To(BeTrue(), "no Hubble flow with %s and %s", source, destination)
		}).Should(Succeed())
	})
}

// expectDynamicLabel waits for a CiliumIdentity having the security label "k8s:<labelKey>=<value>",
// and checks that its identity_info series has the corresponding "<metricName>=<value>" label.
// It returns the numeric identity.
func expectDynamicLabel(labelKey, metricName, value string) string {
	GinkgoHelper()
	var identity string
	Eventually(func(g Gomega) {
		idList := kubectlGetSafe[unstructured.UnstructuredList](g, "ciliumid")

		identity = ""
		for _, id := range idList.Items {
			v, ok, err := unstructured.NestedString(id.Object, "security-labels", "k8s:"+labelKey)
			g.Expect(err).NotTo(HaveOccurred())
			if ok && v == value {
				identity = id.GetName()
				break
			}
		}
		g.Expect(identity).NotTo(BeEmpty(), "CiliumIdentity for %s=%s is not created yet", labelKey, value)

		output := string(scrapeClusterLeader(g))
		var line string
		for l := range strings.Lines(output) {
			if strings.HasPrefix(l, fmt.Sprintf(`neco_cluster_ciliumid_identity_info{identity="%s",`, identity)) {
				line = l
				break
			}
		}
		g.Expect(line).NotTo(BeEmpty(), "identity_info not found for identity %s", identity)
		g.Expect(line).To(ContainSubstring(fmt.Sprintf(`,%s="%s",`, metricName, value)))
		g.Expect(line).To(ContainSubstring(`,namespace="default",`))
		g.Expect(line).To(MatchRegexp(`,security_labels="[^"]*k8s:%s=%s[,"]`, regexp.QuoteMeta(labelKey), value))
	}).Should(Succeed())
	return identity
}

// getEndpointIdentity returns the numeric identity of the CiliumEndpoint of a Pod matching the selector.
func getEndpointIdentity(g Gomega, namespace, selector string) string {
	pods := kubectlGetSafe[corev1.PodList](g, "pod", "-n="+namespace, "-l="+selector)
	g.Expect(pods.Items).NotTo(BeEmpty())

	cep := kubectlGetSafe[unstructured.Unstructured](g, "ciliumendpoint", "-n="+namespace, pods.Items[0].Name)
	id, ok, err := unstructured.NestedInt64(cep.Object, "status", "identity", "id")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ok).To(BeTrue(), "identity is not set for CiliumEndpoint %s/%s", namespace, pods.Items[0].Name)
	return strconv.FormatInt(id, 10)
}

// findSecurityLabels returns the security_labels of the identity_info series for the identity.
func findSecurityLabels(g Gomega, output, identity string) string {
	re := regexp.MustCompile(fmt.Sprintf(`^neco_cluster_ciliumid_identity_info\{identity="%s",(?:.*,)?security_labels="([^"]*)",`, identity))
	for line := range strings.Lines(output) {
		if m := re.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	g.Expect(false).To(BeTrue(), "identity_info not found for identity %s", identity)
	return ""
}
