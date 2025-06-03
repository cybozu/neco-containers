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
		bf, _ := setBmcParam("../config/idrac-test-config.json")
		ctx := context.Background()
		var b Bmc
		var err error
		var job *url.URL
		downloadDir, _ := os.Getwd()
		filename := filepath.Join(downloadDir, "test-tsr.zip")

		It("Create new iDRAC Redfish endpoint", func() {
			b, err = NewBmcEp(bf.IpV4, bf.User, bf.Pass)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Request iDRAC to create TSR", func() {
			job, err = b.StartCollection(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Waiting JOB to collect TSR in iDRAC", func() {
			err = b.WaitCollection(ctx, job)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Download TSR from iDRAC", func() {
			f, err := os.Create(filename)
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			err = b.DownloadSupportAssist(ctx, f)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Remove TSR file", func() {
			err := os.Remove(filename)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
