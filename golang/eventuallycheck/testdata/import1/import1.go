package import1

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func testEventually2() {
	ginkgo.It("should execute eventually", func() {
		gomega.Eventually(func() error {
			return nil
		}).Should(gomega.Succeed())
	})

	ginkgo.It("should not execute eventually", func() {
		gomega.Eventually(func() error {
			return nil
		})
	})
}
