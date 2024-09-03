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
			ml, err := readMachineList("testdata/configmap/machines-list-test.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(ml[0].Serial).To(Equal("server1"))
			Expect(ml[0].BmcIP).To(Equal("192.168.0.1"))
			Expect(ml[0].NodeIP).To(Equal("172.16.0.1"))

			Expect(ml[1].Serial).To(Equal("server2"))
			Expect(ml[1].BmcIP).To(Equal("192.168.0.2"))
			Expect(ml[1].NodeIP).To(Equal("172.16.0.2"))

			Expect(ml[2].Serial).To(Equal("server3"))
			Expect(ml[2].BmcIP).To(Equal("192.168.0.3"))
			Expect(ml[2].NodeIP).To(Equal("172.16.0.3"))

			Expect(ml[3].Serial).To(Equal("server4"))
			Expect(ml[3].BmcIP).To(Equal("192.168.0.4"))
			Expect(ml[3].NodeIP).To(Equal("172.16.0.4"))

			Expect(ml[4].Serial).To(Equal("server5"))
			Expect(ml[4].BmcIP).To(Equal("192.168.0.5"))
			Expect(ml[4].NodeIP).To(Equal("172.16.0.5"))
		})
	})

	Context("Abnormal", func() {
		It("Abnormal, no existing file", func() {
			_, err := readMachineList("testdata/configmap/noexist.json")
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal, lack of element", func() {
			_, err := readMachineList("testdata/configmap/damaged.json")
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal, read empty JSON file", func() {
			_, err := readMachineList("testdata/configmap/empty.json")
			Expect(err).To(HaveOccurred())
		})
	})
})
