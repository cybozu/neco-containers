package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

/*
Test the behavior of bmcClient() accessing iDRAC internal web services
*/
var _ = Describe("Access BMC", Ordered, func() {
	BeforeAll(func() {
		fmt.Println("BeforeAll")
		start_iDRAC_Simulator_ut()
		time.Sleep(10 * time.Second)
	})

	Context("Access iDRAC server to get SEL", func() {
		It("Normal access", func() {
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
			url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, err := bmcClient(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
			url := "https://127.0.0.1:8090/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, err := bmcClient(url)
			Expect(err).To(HaveOccurred())
			errmsg := fmt.Sprintf("Get \"%s\": dial tcp 127.0.0.1:8090: connect: connection refused", url)
			Expect(err.Error()).To(Equal(errmsg))
			Expect(len(byteJSON)).To(Equal(0))
		})

		It("Abnormal access, wrong path", func() {
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
			url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/Entries"
			_, err := bmcClient(url)
			Expect(err).To(HaveOccurred())
			fmt.Println("Err message = ", err)
		})

		It("Abnormal access, wrong password", func() {
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "bad_pass")
			url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			_, err := bmcClient(url)
			Expect(err).To(HaveOccurred())
			fmt.Println("Err message = ", err)
		})

	})
	AfterAll(func() {
		fmt.Println("Shutdown webserver")
		// dummy
	})
})
