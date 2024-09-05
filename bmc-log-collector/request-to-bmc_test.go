package main

import (
	"context"
	"crypto/tls"
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
	username := "support"
	password := basicAuthPassword
	client := &http.Client{
		Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 30 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).DialContext,
		},
	}

	BeforeAll(func() {
		GinkgoWriter.Println("*** Start iDRAC Stub")
		bm1 := bmcMock{
			host:          "127.0.0.1:19082",
			resDir:        "testdata/redfish_response",
			files:         []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
			accessCounter: make(map[string]int),
			responseFiles: make(map[string][]string),
			responseDir:   make(map[string]string),
			isInitmap:     true,
		}
		bm1.startMock()
		// Wait for starting mock web server
		time.Sleep(10 * time.Second)
	})
	AfterAll(func() {
		time.Sleep(3 * time.Second)
		cancel()
	})

	Context("Access iDRAC server to get SEL", func() {
		It("Normal access", func() {
			url := "https://127.0.0.1:19082/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			byteJSON, httpStatusCode, err := requestToBmc(ctx, username, password, client, url)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpStatusCode).To(Equal(200))
			Expect(len(byteJSON)).To(Equal(776))
		})

		It("Abnormal access, not existing web server", func() {
			bad_url := "https://127.0.0.9:19082/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			_, _, err := requestToBmc(ctx, username, password, client, bad_url)
			Expect(err).To(HaveOccurred())
		})

		It("Abnormal access, wrong path", func() {
			wrong_path := "https://127.0.0.1:19082/redfish/v1/Managers/iDRAC.Embedded.1/LogServ1ces/Sel/EntriesWrong"
			_, httpStatusCode, err := requestToBmc(ctx, username, password, client, wrong_path)
			Expect(httpStatusCode).To(Equal(404))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Abnormal access, wrong username", func() {
			url := "https://127.0.0.1:19082/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			bad_username := "badname"
			_, httpStatusCode, err := requestToBmc(ctx, bad_username, password, client, url)
			Expect(err).ToNot(HaveOccurred())
			Expect(httpStatusCode).To(Equal(401))
		})

		It("Abnormal access, wrong password", func() {
			url := "https://127.0.0.1:19082/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
			bad_password := "badpassword"
			_, httpStatusCode, err := requestToBmc(ctx, username, bad_password, client, url)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpStatusCode).To(Equal(401))
		})
	})
})
