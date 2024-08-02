package main

import (
	"context"
	"crypto/tls"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net"
	"net/http"
	"time"
)

/*
Test the behavior of accessing iDRAC internal web services
*/
var _ = Describe("Access BMC", Ordered, func() {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	client := &http.Client{
		Timeout: time.Duration(10) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 20 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
		},
	}
	rfc := RedfishClient{
		user:     "user",
		password: "pass",
		client:   client,
	}

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

	Context("Access iDRAC server to get SEL", func() {
		var redfish_url = "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
		It("Normal access", func() {
			byteJSON, err := requestToBmc(ctx, redfish_url, rfc)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			test_url := "https://127.0.0.9:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, err := requestToBmc(ctx, test_url, rfc)
			Expect(err).To(HaveOccurred())
			errmsg := fmt.Sprintf("Get \"%s\": dial tcp 127.0.0.9:8080: connect: connection refused", test_url)
			Expect(err.Error()).To(Equal(errmsg))
			Expect(len(byteJSON)).To(Equal(0))
		})

		It("Abnormal access, wrong path", func() {
			wrong_url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/EntriesWrong"
			_, err := requestToBmc(ctx, wrong_url, rfc)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong username", func() {
			rfc.user = "badname"
			_, err := requestToBmc(ctx, redfish_url, rfc)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong password", func() {
			rfc.password = "badpw"
			_, err := requestToBmc(ctx, redfish_url, rfc)
			Expect(err).To(HaveOccurred())
		})
	})

	AfterAll(func() {
		GinkgoWriter.Println("shutdown BMC stub")
		client.CloseIdleConnections()
		cancel()
	})
})
