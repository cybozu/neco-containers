package bmc

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get User from bmc-user.json", Ordered, func() {
	Context("Normal", func() {
		It("Read JSON file", func() {
			user, err := ReadUsers("../testdata/bmc-user.json")
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("User", user.Support)
			Expect(user.Support.Password.Raw).To(Equal("raw password for support user"))
		})
	})
	Context("Abnormal", func() {
		It("Read no existing file", func() {
			_, err := ReadUsers("../testdata/no-exist.json")
			Expect(err).To(HaveOccurred())
		})
		It("no support user in json file", func() {
			_, err := ReadUsers("../testdata/bmc-user-err.json")
			Expect(err).To(HaveOccurred())
		})
	})
})
