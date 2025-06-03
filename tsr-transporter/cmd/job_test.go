package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/cybozu/neco-containers/tsr-transporter/bmc"
	"github.com/cybozu/neco-containers/tsr-transporter/kintone"
	"github.com/cybozu/neco-containers/tsr-transporter/sabakan"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TSR Transporter", Ordered, func() {
	BeforeAll(func(ctx SpecContext) {
		GinkgoWriter.Println("Start stub servers")
		saba := sabakanMock{
			host:   "127.0.0.1:7180",
			path:   "/api/v1/machines",
			resDir: "../testdata/sabakan-data",
		}
		saba.startMock()
		By("Wait for mock server become up: " + saba.getEndpoint())
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", saba.getEndpoint(), nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())
	}, NodeTimeout(10*time.Second))

	var b *bmc.UserConfig
	var s *sabakan.Config
	var k *kintone.Config
	var err error

	Context("Config files access test", func() {
		It("Read BMC config file", func() {
			b, err = bmc.ReadUsers("../config/bmc-user.json")
			Expect(err).ToNot(HaveOccurred())
		})
		It("Sabakana config file", func() {
			s, err = sabakan.ReadConfig("../testdata/sabakan.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(s.Ep).To(Equal("http://127.0.0.1:7180/api/v1/machines"))
		})
		It("Kintone config file", func() {
			k, err = kintone.ReadAppConfig("../config/kintone-test-config.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(k.Domain).To(Equal("https://6hu5ta9d6e4z.cybozu.com"))
		})
		It("TSR-transporter Job test", func() {
			ctx := context.Background()
			err = jobMain(ctx, b, s, k)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
