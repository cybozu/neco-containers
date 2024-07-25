package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
Tests machineListReader(), which reads a CSV file with a specified path and sets it into a structure.
*/
var _ = Describe("Get Machines List", Ordered, func() {
	Context("Normal", func() {
		It("Read JSON file", func() {
			ml, err := machineListReader("testdata/configmap/machinelist-test.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(ml.Machine[0].Serial).To(Equal("server1"))
			Expect(ml.Machine[0].BmcIP).To(Equal("192.168.0.1"))
			Expect(ml.Machine[0].NodeIP).To(Equal("172.16.0.1"))
			Expect(ml.Machine[4].Serial).To(Equal("server5"))
			Expect(ml.Machine[4].BmcIP).To(Equal("192.168.0.5"))
			Expect(ml.Machine[4].NodeIP).To(Equal("172.16.0.5"))
		})
	})

	Context("Abnormal", func() {

		It("Abnormal, no existing file", func() {
			_, err := machineListReader("testdata/configmap/noexist.json")
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal, lack of element", func() {
			_, err := machineListReader("testdata/configmap/damaged.json")
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal, read empty JSON file", func() {
			_, err := machineListReader("testdata/configmap/empty.json")
			Expect(err).To(HaveOccurred())
		})

	})
})
