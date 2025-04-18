package dell

import (
	"context"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dell iDRAC access Interface Library", func() {
	Context("basic API test", Ordered, func() {
		bf, _ := setBmcParam("../local/idrac-test-config.json")
		ctx := context.Background()
		var bmc Bmc
		var err error
		var job *url.URL

		It("Create new iDRAC Redfish endpoint", func() {
			bmc, err = NewBmcEp(bf.IpV4, bf.User, bf.Pass)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Request iDRAC to create TSR", func() {
			job, err = bmc.StartCollection(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Waiting JOB to collect TSR in iDRAC", func() {
			err = bmc.WaitCollection(ctx, job)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Download TSR from iDRAC", func() {
			downloadDir, _ := os.Getwd()
			filename := filepath.Join(downloadDir, "test-tsr.zip")
			f, err := os.Create(filename)
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			err = bmc.DownloadSupportAssist(ctx, f)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
