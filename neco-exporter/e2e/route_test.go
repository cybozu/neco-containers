package e2e

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

const routeHelperPod = "route-helper"

func getRoutePod(g Gomega) string {
	pods := kubectlGetSafe[corev1.PodList](g, "pod", nsOption, "-l=app.kubernetes.io/name=neco-node-route-exporter")
	g.Expect(pods.Items).To(HaveLen(1))
	return pods.Items[0].Name
}

// scrapeRoute scrapes the route collector, which serves on the host network.
func scrapeRoute(g Gomega) []byte {
	pods := kubectlGetSafe[corev1.PodList](g, "pod", nsOption, "-l=app.kubernetes.io/name=neco-node-route-exporter")
	g.Expect(pods.Items).To(HaveLen(1))
	return scrape(g, pods.Items[0].Status.HostIP+":8090")
}

// routeLogs returns the JSON logs of the route collector with the message.
func routeLogs(g Gomega, msg string) []map[string]any {
	stdout := kubectlSafe(g, nil, "logs", nsOption, getRoutePod(g))
	var ret []map[string]any
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		var e map[string]any
		if json.Unmarshal(sc.Bytes(), &e) != nil {
			continue
		}
		if e["msg"] == msg {
			ret = append(ret, e)
		}
	}
	return ret
}

// ensureRouteHelper creates the helper Pod to change the routes of the node if it does not exist.
// The Pod is shared by the specs and deleted with the cluster.
func ensureRouteHelper() {
	Eventually(func(g Gomega) {
		kubectlSafe(g, nil, "apply", "-f", "testdata/route-helper-pod.yaml")
		kubectlSafe(g, nil, nsOption, "wait", "--for=condition=Ready", "pod/"+routeHelperPod, "--timeout=10s")
	}).Should(Succeed())
}

// logDelay returns the time from not_before to the time of the log.
func logDelay(g Gomega, e map[string]any) time.Duration {
	g.Expect(e).To(And(HaveKey("time"), HaveKey("not_before")))
	logged, err := time.Parse(time.RFC3339Nano, e["time"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	notBefore, err := time.Parse(time.RFC3339Nano, e["not_before"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	return logged.Sub(notBefore)
}

// onNode runs a shell script in the host network namespace of the node.
func onNode(g Gomega, script string) {
	kubectlSafe(g, nil, "exec", nsOption, routeHelperPod, "--", "sh", "-c", script)
}

func testRouteCollector() {
	// The test routes have larger metrics than the default route of the kind node, so they do not change the routing.
	const (
		testProto = "bird"

		testMetric  = "4242"
		routeParams = "default dev lo proto " + testProto + " metric " + testMetric

		dummyLink   = "route-test"
		dummyMetric = "4243"
	)

	It("should report the default route of the node", func() {
		Eventually(func(g Gomega) {
			output := scrapeRoute(g)

			health, ok := findMetricValue(g, output, `neco_node_collector_health{collector="route"} `)
			g.Expect(ok).To(BeTrue(), "health metric not found")
			g.Expect(health).To(BeNumerically("==", 1))

			present, ok := findMetricValue(g, output, fmt.Sprintf(`neco_node_route_default_present{node="%s"} `, getNodeName(g)))
			g.Expect(ok).To(BeTrue(), "default_present not found")
			g.Expect(present).To(BeNumerically("==", 1))
		}).Should(Succeed())
	})

	It("should record a default route added and deleted", func() {
		ensureRouteHelper()
		DeferCleanup(func() {
			_, _, _ = kubectl(nil, "exec", nsOption, routeHelperPod, "--", "sh", "-c", "ip route del "+routeParams+" || true")
		})

		By("adding a default route")
		Eventually(func(g Gomega) {
			onNode(g, "ip route replace "+routeParams)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := scrapeRoute(g)
			node := getNodeName(g)

			nexthops, ok := findMetricValue(g, output,
				fmt.Sprintf(`neco_node_route_default_nexthops{metric="%s",node="%s"} `, testMetric, node))
			g.Expect(ok).To(BeTrue(), "default_nexthops of the test route not found")
			g.Expect(nexthops).To(BeNumerically("==", 1))
		}).Should(Succeed())

		By("deleting the default route")
		Eventually(func(g Gomega) {
			onNode(g, "ip route del "+routeParams)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := scrapeRoute(g)
			node := getNodeName(g)

			deleted, ok := findMetricValue(g, output,
				fmt.Sprintf(`neco_node_route_changes_total{default="true",node="%s",op="delete",protocol="%s"} `, node, testProto))
			g.Expect(ok).To(BeTrue(), "changes_total of the deletion not found")
			g.Expect(deleted).To(BeNumerically(">=", 1))

			_, ok = findMetricValue(g, output,
				fmt.Sprintf(`neco_node_route_default_nexthops{metric="%s",node="%s"} `, testMetric, node))
			g.Expect(ok).To(BeFalse(), "default_nexthops of the deleted route still exists")

			// The node keeps its own default route.
			nodePresent, ok := findMetricValue(g, output, fmt.Sprintf(`neco_node_route_default_present{node="%s"} `, node))
			g.Expect(ok).To(BeTrue(), "default_present not found")
			g.Expect(nodePresent).To(BeNumerically("==", 1))
		}).Should(Succeed())
	})

	It("should notice next hops and routes removed without notifications", func() {
		ensureRouteHelper()
		links := []string{dummyLink + "0", dummyLink + "1"}
		DeferCleanup(func() {
			for _, l := range links {
				_, _, _ = kubectl(nil, "exec", nsOption, routeHelperPod, "--", "ip", "link", "del", l)
			}
		})
		nexthopsPrefix := func(g Gomega) string {
			return fmt.Sprintf(`neco_node_route_default_nexthops{metric="%s",node="%s"} `, dummyMetric, getNodeName(g))
		}
		changesPrefix := func(g Gomega) string {
			return fmt.Sprintf(`neco_node_route_changes_total{default="true",node="%s",op="delete",protocol="%s"} `, getNodeName(g), testProto)
		}

		By("adding a multipath default route through two dummy links")
		Eventually(func(g Gomega) {
			for _, l := range links {
				onNode(g, fmt.Sprintf("{ ip link show %[1]s >/dev/null 2>&1 || ip link add %[1]s type dummy; } && ip link set %[1]s up", l))
			}
			onNode(g, fmt.Sprintf("ip route replace default proto %s metric %s nexthop dev %s nexthop dev %s",
				testProto, dummyMetric, links[0], links[1]))
		}).Should(Succeed())

		var deletedBefore float64
		Eventually(func(g Gomega) {
			output := scrapeRoute(g)
			nexthops, ok := findMetricValue(g, output, nexthopsPrefix(g))
			g.Expect(ok).To(BeTrue(), "default_nexthops of the test route not found")
			g.Expect(nexthops).To(BeNumerically("==", 2))
			deletedBefore, _ = findMetricValue(g, output, changesPrefix(g))
		}).Should(Succeed())

		By("setting one of the links down, which marks the next hop dead without a notification")
		Eventually(func(g Gomega) {
			onNode(g, "ip link set "+links[1]+" down")
		}).Should(Succeed())

		// The periodic dump finds the dead next hop.
		Eventually(func(g Gomega) {
			nexthops, ok := findMetricValue(g, scrapeRoute(g), nexthopsPrefix(g))
			g.Expect(ok).To(BeTrue(), "default_nexthops of the test route not found")
			g.Expect(nexthops).To(BeNumerically("==", 1))
		}).Should(Succeed())

		By("setting the other link down, which removes the route without a notification")
		Eventually(func(g Gomega) {
			onNode(g, "ip link set "+links[0]+" down")
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := scrapeRoute(g)
			_, ok := findMetricValue(g, output, nexthopsPrefix(g))
			g.Expect(ok).To(BeFalse(), "default_nexthops of the removed route still exists")

			deleted, _ := findMetricValue(g, output, changesPrefix(g))
			g.Expect(deleted).To(BeNumerically("==", deletedBefore), "the removal was notified")
		}).Should(Succeed())
	})

	It("should log when the node loses its default route and when it comes back", func() {
		ensureRouteHelper()
		By("deleting the default route of the node for a few seconds")
		// The script restores the route even if it fails in the middle, as the node needs it.
		Eventually(func(g Gomega) {
			onNode(g, `set -e
saved=$(ip -4 route show default table main)
[ "$(echo "$saved" | wc -l)" -eq 1 ]
trap 'ip route replace $saved' EXIT
ip route del $saved
sleep 3`)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			lost := routeLogs(g, "default route is missing")
			g.Expect(lost).NotTo(BeEmpty(), "the loss is not logged")
			l := lost[len(lost)-1]
			g.Expect(l).To(HaveKeyWithValue("level", "WARN"))
			g.Expect(l).To(HaveKeyWithValue("detected_by", "notification"))
			// The notification triggers a dump soon, so the loss is logged right after it happened.
			g.Expect(logDelay(g, l)).To(BeNumerically("<", time.Second))

			// Every notification of default routes is logged.
			changes := routeLogs(g, "default route changed")
			g.Expect(changes).To(ContainElement(And(HaveKeyWithValue("op", "delete"), HaveKeyWithValue("metric", BeNumerically("==", 0)))))
			g.Expect(changes).To(ContainElement(And(HaveKeyWithValue("op", "add"), HaveKeyWithValue("metric", BeNumerically("==", 0)))))

			restored := routeLogs(g, "default route is restored")
			g.Expect(restored).NotTo(BeEmpty(), "the restoration is not logged")
			e := restored[len(restored)-1]
			g.Expect(e).To(HaveKeyWithValue("level", "INFO"))
			g.Expect(e).To(HaveKeyWithValue("detected_by", "notification"))
			g.Expect(logDelay(g, e)).To(BeNumerically("<", time.Second))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			output := scrapeRoute(g)
			node := getNodeName(g)

			present, ok := findMetricValue(g, output, fmt.Sprintf(`neco_node_route_default_present{node="%s"} `, node))
			g.Expect(ok).To(BeTrue(), "default_present not found")
			g.Expect(present).To(BeNumerically("==", 1))
		}).Should(Succeed())
	})
}
