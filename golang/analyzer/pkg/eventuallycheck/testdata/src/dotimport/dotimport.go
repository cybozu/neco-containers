package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testEventually() {
	It("should execute Assert functions", func() {
		Consistently(func() error {
			return nil
		}).Should(Succeed())
		ConsistentlyWithOffset(1, func() error {
			return nil
		}).Should(Succeed())
		Eventually(func() error {
			return nil
		}).Should(Succeed())
		EventuallyWithOffset(1, func() error {
			return nil
		}).Should(Succeed())
		Expect(true).To(BeTrue())
		ExpectWithOffset(1, true).To(BeTrue())
		Ω(true).To(BeTrue())
	})

	It("should not execute eventually", func() {
		Eventually(func() error { // want "invalid Assertion: Should/ShouldNot not called"
			return nil
		})
		Consistently(func() error { // want "invalid Assertion: Should/ShouldNot not called"
			return nil
		})
		Expect(true)              // want "invalid Assertion: Should/ShouldNot not called"
		ExpectWithOffset(1, true) // want "invalid Assertion: Should/ShouldNot not called"
		Ω(true)                   // want "invalid Assertion: Should/ShouldNot not called"
	})
}

func main() {
	testEventually()
}
