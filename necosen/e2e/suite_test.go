package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/cybozu-go/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test")
}

var _ = BeforeSuite(func() {
	fmt.Println("Preparing...")

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(5 * time.Minute)

	log.DefaultLogger().SetOutput(GinkgoWriter)

	fmt.Println("Begin tests...")
})

var _ = Describe("Test necosen", func() {
	BeforeEach(func() {
		fmt.Printf("START: %s\n", time.Now().Format(time.RFC3339))
	})
	AfterEach(func() {
		fmt.Printf("END: %s\n", time.Now().Format(time.RFC3339))
	})

	runTest()
})

func testSetup() {
	var validClientPodIP string
	var validClientPodName, invalidClientPodName string
	It("should retrieve the valid client's IP", func() {
		stdout := kubectlSafe(nil, "get", "cm", "-n=contour", "necosen-config", "-o=json")
		stdout = yqSafe(stdout, `.data["config.yaml"]`)
		stdout = yqSafe(stdout, `.sourceIP.allowedCIDRs[0] | split("/") | .[0]`)
		validClientPodIP = string(stdout)

		stdout = kubectlSafe(nil, "get", "pod", "--field-selector=status.podIP="+validClientPodIP, "-o=jsonpath={.items[0].metadata.name}")
		validClientPodName = string(stdout)
	})

	It("should retrieve the invalid client's IP", func() {
		stdout := kubectlSafe(nil, "get", "pod", "-l=app.kubernetes.io/name=client", "--field-selector=status.podIP!="+validClientPodIP, "-o=jsonpath={.items[0].metadata.name}")
		invalidClientPodName = string(stdout)
		fmt.Println("1:" + string(stdout))
		fmt.Println(invalidClientPodName)
	})

	var envoyIP string
	It("should get the envoy's IP", func() {
		stdout := kubectlSafe(nil, "get", "svc", "-n=contour", "contour-envoy", "-o=jsonpath={.spec.clusterIP}")
		envoyIP = string(stdout)
	})

	It("should allow access from the valid client to the default-allow service", func() {
		stdout := kubectlSafe(nil, "exec", validClientPodName, "--", "curl", "-sk", "--resolv", "frontend-free.example.com:443:"+envoyIP, "https://frontend-free.example.com/")
		Expect(string(stdout)).To(Equal("Hello"))
	})

	It("should allow access from the invalid client to the default-allow service", func() {
		stdout := kubectlSafe(nil, "exec", invalidClientPodName, "--", "curl", "-sk", "--resolv", "frontend-free.example.com:443:"+envoyIP, "https://frontend-free.example.com/")
		Expect(string(stdout)).To(Equal("Hello"))
	})

	It("should allow access from the valid client", func() {
		stdout := kubectlSafe(nil, "exec", validClientPodName, "--", "curl", "-sk", "--resolv", "frontend.example.com:443:"+envoyIP, "https://frontend.example.com/")
		Expect(string(stdout)).To(Equal("Hello"))
	})

	It("should deny access from the invalid client", func() {
		stdout := kubectlSafe(nil, "exec", invalidClientPodName, "--", "curl", "-sk", "--resolv", "frontend.example.com:443:"+envoyIP, "https://frontend.example.com/")
		Expect(string(stdout)).NotTo(Equal("Hello"))
	})
}

func runTest() {
	Context("setup clusters", testSetup)
}
