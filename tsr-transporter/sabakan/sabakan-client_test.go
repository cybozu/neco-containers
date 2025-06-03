package sabakan

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sabakan Interface Library", func() {
	var _ = Describe("Sabakan mock", Ordered, func() {
		saba := sabakanMock{
			host:   "127.0.0.1:7180",
			path:   "/api/v1/machines",
			resDir: "../testdata/sabakan-data",
		}
		BeforeAll(func(ctx SpecContext) {
			saba.startMock()
			By("Wait for mock server become up: " + saba.getEndpoint())
			Eventually(func(ctx SpecContext) error {
				req, _ := http.NewRequest("GET", saba.getEndpoint(), nil)
				client := &http.Client{Timeout: time.Duration(3) * time.Second}
				_, err := client.Do(req)
				return err
			}).WithContext(ctx).Should(Succeed())
		}, NodeTimeout(10*time.Second))

		Context("Test Sabakan access library", func() {
			It("Read config of sabakan access pointt", func(ctx SpecContext) {
				saba, err := ReadConfig("../testdata/sabakan.json")
				Expect(err).ToNot(HaveOccurred())
				Expect(saba.Service).To(Equal("127.0.0.1:7180"))
				Expect(saba.Path).To(Equal("/api/v1/machines"))
				Expect(saba.Ep).To(Equal("http://127.0.0.1:7180/api/v1/machines"))
			}, SpecTimeout(3*time.Second))

			It("Get IPv4 from Serial", func(ctx SpecContext) {
				ipv4, err := GetBmcIpv4(saba.getEndpoint(), "1PNKVQ3")
				Expect(err).ToNot(HaveOccurred())
				Expect(ipv4).To(Equal("10.72.17.6"))
			}, SpecTimeout(3*time.Second))

			It("Confirm behavior when the serial does not exist", func(ctx SpecContext) {
				ipv4, err := GetBmcIpv4(saba.getEndpoint(), "UNEXIST")
				Expect(err).ToNot(HaveOccurred())
				Expect(ipv4).To(Equal(""))
			}, SpecTimeout(3*time.Second))
		})
	})
})
