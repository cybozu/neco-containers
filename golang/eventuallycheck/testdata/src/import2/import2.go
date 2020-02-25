package main

import (
	gin "github.com/onsi/ginkgo"
	gome "github.com/onsi/gomega"
)

func testEventually2() {
	gin.It("should execute eventually", func() {
		gome.Eventually(func() error {
			return nil
		}).Should(gome.Succeed())
	})

	gin.It("should not execute eventually", func() {
		gome.Eventually(func() error { // want "invalid Eventually: Assertion not called"
			return nil
		})
	})
}

func main() {
	testEventually2()
}
