package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testEventually() {
	It("should execute eventually", func() {
		Eventually(func() error {
			return nil
		}).Should(Succeed())
	})

	It("should not execute eventually", func() {
		Eventually(func() error { // want "invalid Eventually: Assertion not called"
			return nil
		})
	})
}

func main() {
	testEventually()
}
