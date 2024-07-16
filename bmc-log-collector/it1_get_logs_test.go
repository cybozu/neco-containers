package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	ctx, cancel := context.WithCancel(context.Background())
	redfishPath := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	var mu sync.Mutex

	// Start iDRAC Simulator
	BeforeAll(func() {
		fmt.Println("Start iDRAC Simulator IT-1")
		startIdracMock_it(&mu)
		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("IT read maachine list CSV file", func() {
		var machinesList Machines
		var err error
		It("Read CSV file", func() {
			machinesList, err = machineListReader("testdata/bmc-list-it.csv")
			Expect(err).NotTo(HaveOccurred())
			//GinkgoWriter.Println(machines_list)
			fmt.Println("Read CSV file", machinesList)

		})

		It("Put que", func(ctx SpecContext) {
			//GinkgoWriter.Println(machines_list.machine)
			fmt.Println(machinesList.machine)
			putQueue(machinesList.machine)
		}, SpecTimeout(time.Second))

		// Start log collector
		It("Print iDRAC log for loki 1", func() {
			// Get target BMC from queue
			v := getQueue()
			//GinkgoWriter.Println("Get BMC-IP(", v.BmcIP, ") from queue")
			fmt.Println("Get BMC-IP(", v.BmcIP, ") from queue")
			redfishUrl := "https://" + v.BmcIP + redfishPath
			//GinkgoWriter.Println("URL = ", redfish_url)
			fmt.Println("URL = ", redfishUrl)
			// Get SEL from iDRAC
			byteData, err := bmcClient(redfishUrl)
			//GinkgoWriter.Println(string(byteData))
			fmt.Println(string(byteData))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Put queue again for test", func() {
			//GinkgoWriter.Println("Put que ==", machines_list.machine)
			fmt.Println("Put que ==", machinesList.machine)
			// reset access counter
			access_counter[redfishPath] = 0
			putQueue(machinesList.machine)
		})

		It("Print iDRAC log for loki 2 ", func() {
			GinkgoWriter.Println("=== 2ND ACCESS ===")
			go collector(ctx, 1, redfishPath, "pointers_it1")
			time.Sleep(20 * time.Second)
		})
	})
	AfterAll(func() {
		fmt.Println("Shutdown iDRAC Simulator IT-1")
		cancel()
		time.Sleep(5 * time.Second)
	})
})
