package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
Tests machineListReader(), which reads a CSV file with a specified path and sets it into a structure.
*/
var _ = Describe("Get Machines List", Ordered, func() {
	BeforeAll(func() {
		fmt.Println("Get Machines List")
	})

	Context("Normal CSV file", func() {
		It("Read CSV file", func() {
			ml, err := machineListReader("testdata/bmc-list-ut.csv")
			Expect(err).NotTo(HaveOccurred())
			Expect(ml.machine[0].Serial).To(Equal("server1"))
			Expect(ml.machine[0].BmcIP).To(Equal("192.168.0.1"))
			Expect(ml.machine[0].NodeIP).To(Equal("172.16.0.1"))
			Expect(ml.machine[4].Serial).To(Equal("server5"))
			Expect(ml.machine[4].BmcIP).To(Equal("192.168.0.5"))
			Expect(ml.machine[4].NodeIP).To(Equal("172.16.0.5"))
		})
		It("Abnormal, no existing file", func() {
			_, err := machineListReader("testdata/noexist.csv")
			Expect(err).To(HaveOccurred())
		})
		It("Abnormal, read empty CSV file", func() {
			_, err := machineListReader("testdata/empty.csv")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
