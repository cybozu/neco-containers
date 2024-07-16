package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"sync"
	"time"
)

/*
Test the behavior of bmcClient() accessing iDRAC internal web services
*/
var _ = Describe("Access BMC", Ordered, func() {
	var mu sync.Mutex

	BeforeAll(func() {
		fmt.Println("*** Start iDRAC Simulator UT-1")
		start_iDRAC_Simulator_ut(&mu)
		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	var redfish_url = "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"

	Context("Access iDRAC server to get SEL", func() {
		It("Normal access", func() {
			byteJSON, err := bmcClient(redfish_url)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			test_url := "https://127.0.0.1:8090/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, err := bmcClient(test_url)
			Expect(err).To(HaveOccurred())
			errmsg := fmt.Sprintf("Get \"%s\": dial tcp 127.0.0.1:8090: connect: connection refused", test_url)
			Expect(err.Error()).To(Equal(errmsg))
			Expect(len(byteJSON)).To(Equal(0))
		})

		It("Abnormal access, wrong path", func() {
			wrong_url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/EntriesWrong"
			_, err := bmcClient(wrong_url)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong password", func() {
			os.Setenv("BMC_PASS", "bad_pass")
			_, err := bmcClient(redfish_url)
			Expect(err).To(HaveOccurred())
		})

	})
	AfterAll(func() {
		fmt.Println("*** Shutdown iDRAC Simulator UT-1")
	})
})
