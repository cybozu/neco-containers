package sabakan

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sabakan Interface Library", func() {
	var _ = Describe("sabakan mock", Ordered, func() {
		saba := sabakanMock{
			host:   "127.0.0.1:7180",
			path:   "/api/v1/machines",
			resDir: "../testdata/sabakan",
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

		Context("Test GetBmcIpv4", func() {
			It("test", func(ctx SpecContext) {
				ipv4, err := GetBmcIpv4(saba.getEndpoint(), "1PNKVQ3")
				Expect(err).ToNot(HaveOccurred())
				fmt.Println("ipv4", ipv4)
				Expect(ipv4).To(Equal("10.72.17.6"))
			}, SpecTimeout(3*time.Second))

			It("test", func(ctx SpecContext) {
				ipv4, err := GetBmcIpv4(saba.getEndpoint(), "UNEXIST")
				Expect(err).To(HaveOccurred())
				Expect(ipv4).To(Equal(""))
			}, SpecTimeout(3*time.Second))
		})
	})
})
