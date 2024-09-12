package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get User from bmc-user.json", Ordered, func() {
	Context("Normal", func() {
		It("Read JSON file", func() {
			user, err := LoadBMCUserConfig("testdata/etc/bmc-user.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Support.Password.Raw).To(Equal(basicAuthPassword))
		})
	})

	Context("Abnormal", func() {
		It("Read no existing file", func() {
			_, err := LoadBMCUserConfig("testdata/etc/no-exist.json")
			Expect(err).To(HaveOccurred())
		})

		It("no support user in json file", func() {
			_, err := LoadBMCUserConfig("testdata/etc/bmc-user-err.json")
			Expect(err).To(HaveOccurred())
		})
	})
})
