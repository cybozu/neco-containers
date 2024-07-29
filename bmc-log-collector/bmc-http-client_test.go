package main

import (
	"crypto/tls"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	//"os"
	"time"
)

/*
Test the behavior of accessing iDRAC internal web services
*/
var _ = Describe("Access BMC", Ordered, func() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   time.Duration(10) * time.Second,
		Transport: tr,
	}
	rfc := RedfishClient{
		user:     "user",
		password: "pass",
		client:   client,
	}
	//var mu sync.Mutex
	BeforeAll(func() {
		fmt.Println("*** Start iDRAC Stub")
		bm1 := bmcMock{
			host:   "127.0.0.1:8080",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm1.startMock()
		// Wait for starting mock web server
		time.Sleep(10 * time.Second)
	})
	/*
		BeforeEach(func() {
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
		})
	*/

	var redfish_url = "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"

	Context("Access iDRAC server to get SEL", func() {
		It("Normal access", func() {
			byteJSON, err := rfc.requestToBmc(redfish_url)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			test_url := "https://127.0.0.9:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, err := rfc.requestToBmc(test_url)
			Expect(err).To(HaveOccurred())
			errmsg := fmt.Sprintf("Get \"%s\": dial tcp 127.0.0.9:8080: connect: connection refused", test_url)
			Expect(err.Error()).To(Equal(errmsg))
			Expect(len(byteJSON)).To(Equal(0))
		})

		It("Abnormal access, wrong path", func() {
			wrong_url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/EntriesWrong"
			_, err := rfc.requestToBmc(wrong_url)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong username", func() {
			rfc.user = "badname"
			_, err := rfc.requestToBmc(redfish_url)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong password", func() {
			rfc.password = "badpw"
			_, err := rfc.requestToBmc(redfish_url)
			Expect(err).To(HaveOccurred())
		})

	})
	AfterAll(func() {
		fmt.Println("*** Shutdown iDRAC Simulator UT-1")
		client.CloseIdleConnections()
	})
})
