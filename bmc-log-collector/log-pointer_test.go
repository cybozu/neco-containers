package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get Machines List", Ordered, func() {

	var ptr LastPointer
	var err error

	Context("Normal CSV file", func() {
		It("Read ptr file", func() {
			ptr, err = readLastPointer("ABCDEF", "testdata/pointers")
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.Serial).To(Equal("ABCDEF"))
			Expect(ptr.LastReadTime).To(Equal(int64(0)))
			Expect(ptr.LastReadId).To(Equal(0))
			GinkgoWriter.Println(ptr)
		})
		It("Update ptr", func() {
			ptr.LastReadTime = 1
			ptr.LastReadId = 1
			err = updateLastPointer(ptr, "testdata/pointers")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
