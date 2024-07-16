package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	ctx := context.Background()
	ctxParent, cancel := context.WithCancel(ctx)
	redfish_path := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/EntriesIT1"

	// Start iDRAC Simulator
	BeforeAll(func() {
		fmt.Println("Start iDRAC Simulator IT-1")
		start_iDRAC_Simulator_it()
		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("IT read maachine list CSV file", func() {
		var machines_list Machines
		var err error
		It("Read CSV file", func() {
			machines_list, err = machineListReader("testdata/bmc-list-it.csv")
			Expect(err).NotTo(HaveOccurred())
			//GinkgoWriter.Println(machines_list)
			fmt.Println("Read CSV file", machines_list)

		})

		It("Put que", func(ctx SpecContext) {
			//GinkgoWriter.Println(machines_list.machine)
			fmt.Println(machines_list.machine)
			putQueue(machines_list.machine)
		}, SpecTimeout(time.Second))

		// Start log collector
		It("Print iDRAC log for loki 1", func() {
			// Get target BMC from queue
			v := getQueue()
			//GinkgoWriter.Println("Get BMC-IP(", v.BmcIP, ") from queue")
			fmt.Println("Get BMC-IP(", v.BmcIP, ") from queue")
			redfish_url := "https://" + v.BmcIP + redfish_path
			//GinkgoWriter.Println("URL = ", redfish_url)
			fmt.Println("URL = ", redfish_url)
			// Get SEL from iDRAC
			byteData, err := bmcClient(redfish_url)
			//GinkgoWriter.Println(string(byteData))
			fmt.Println(string(byteData))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Put queue again for test", func() {
			//GinkgoWriter.Println("Put que ==", machines_list.machine)
			fmt.Println("Put que ==", machines_list.machine)
			// reset access counter
			access_counter[redfish_path] = 0
			putQueue(machines_list.machine)
		})

		It("Print iDRAC log for loki 2 ", func() {
			GinkgoWriter.Println("=== 2ND ACCESS ===")
			go collector(ctxParent, 1, redfish_path, "pointers_it1")
			time.Sleep(20 * time.Second)
		})
	})
	AfterAll(func() {
		fmt.Println("Shutdown iDRAC Simulator IT-1")
		cancel()
		time.Sleep(5 * time.Second)
	})
})
