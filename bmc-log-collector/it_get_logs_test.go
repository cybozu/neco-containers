package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {
	// Start iDRAC Simulator
	BeforeAll(func() {
		GinkgoWriter.Println("BeforeAll")
		start_iDRAC_Simulator_it()
		time.Sleep(10 * time.Second)
	})

	Context("IT read maachine list CSV file", func() {
		var machines_list Machines
		var err error
		It("Read CSV file", func() {
			machines_list, err = machineListReader("testdata/bmc-list-it.csv")
			Expect(err).NotTo(HaveOccurred())
			GinkgoWriter.Println(machines_list)
		})

		It("Put que", func() {
			GinkgoWriter.Println(machines_list.machine)
			putQueue(machines_list.machine) // エラー処理が無くても良い？
		})

		// Start log collector
		It("Print iDRAC log for loki 1", func() {
			// Get target BMC from queue
			v := getQueue()
			GinkgoWriter.Println("Get BMC-IP(", v.BmcIP, ") from queue")
			redfish_url := "https://" + v.BmcIP + "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries2"
			GinkgoWriter.Println("URL = ", redfish_url)
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
			// Get SEL from iDRAC
			byteData, err := bmcClient(redfish_url)
			GinkgoWriter.Println(string(byteData))
			Expect(err).NotTo(HaveOccurred())
		})

		// もう一度、キューにセットする
		It("Put queue", func() {
			GinkgoWriter.Println("Put que ==", machines_list.machine)
			putQueue(machines_list.machine)
		})

		It("Print iDRAC log for loki 2 ", func() {
			GinkgoWriter.Println("=== 2ND ACCESS ===")
			os.Setenv("BMC_USER", "user")
			os.Setenv("BMC_PASS", "pass")
			collector(0, "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries2")
			// 何を確認する？ 目視ではダメだ
		})

	})
	AfterAll(func() {
		fmt.Println("Shutdown webserver")
		// dummy
	})

})
