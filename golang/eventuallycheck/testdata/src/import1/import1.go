package main

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
		gomega.Eventually(func() error { // want "invalid Eventually: Assertion not called"
			return nil
		})
	})
}

func main()  {
	testEventually2()
}
