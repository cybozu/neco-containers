package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Books", func() {

	var url string

	BeforeEach(func() {
		url = "http://127.0.0.1:2379"
		fmt.Println(url)
	})

	AfterEach(func() {
		fmt.Println("Done")
	})

	Context("Read Machine List CSV file", func() {
		It("Read CSV", func() {
			m, err := MachineListReader("testdata/bmc-list.csv")
			Expect(err).NotTo(HaveOccurred())

			fmt.Println(m)
		})
	})
})
