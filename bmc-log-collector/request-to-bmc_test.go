package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
Test the behavior of accessing iDRAC internal web services
*/
var _ = Describe("Access BMC", Ordered, func() {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	username := "user"
	password := "pass"
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
	url := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"

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
		It("Normal access", func() {
			byteJSON, httpStatusCode, err := requestToBmc(ctx, username, password, client, url)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpStatusCode).To(Equal(200))
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			bad_url := "https://127.0.0.9:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			_, _, err := requestToBmc(ctx, username, password, client, bad_url)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong path", func() {
			wrong_path := "https://127.0.0.1:8080/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/EntriesWrong"
			_, httpStatusCode, err := requestToBmc(ctx, username, password, client, wrong_path)
			Expect(httpStatusCode).To(Equal(404))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Abnormal access, wrong username", func() {
			bad_username := "badname"
			_, httpStatusCode, err := requestToBmc(ctx, bad_username, password, client, url)
			Expect(err).ToNot(HaveOccurred())
			Expect(httpStatusCode).To(Equal(401))
		})

		It("Abnormal access, wrong password", func() {
			bad_password := "badpassword"
			_, httpStatusCode, err := requestToBmc(ctx, username, bad_password, client, url)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpStatusCode).To(Equal(401))
		})
	})

	AfterAll(func() {
		//	GinkgoWriter.Println("shutdown BMC stub")
		//	client.CloseIdleConnections()
		cancel()
	})
})
