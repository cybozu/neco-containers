package e2e

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(5 * time.Minute)

	SetDefaultConsistentlyPollingInterval(time.Second)
	SetDefaultConsistentlyDuration(15 * time.Second)
})

var _ = Describe("Test neco-exporter", func() {
	runTest()
})

func runTest() {
	Context("exporter", testExporter)

	// test cluster collectors
	Context("cert", testCertCollector)
	Context("ciliumid", testCiliumIDCollector)

	// test node collectors
	Context("bpf", testBPFCollector)
}
