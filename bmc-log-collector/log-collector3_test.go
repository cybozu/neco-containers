package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	var wg sync.WaitGroup
	var lc logCollector

	// Start iDRAC Stub
	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		lc = logCollector{
			machinesPath: "testdata/configmap/log-collector-test.json",
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			wg:           &wg,
			testMode:     true,
			testOut:      "testdata/output/",
		}
		GinkgoWriter.Println("Start iDRAC Stub")
		bm := bmcMock{
			host:   "127.0.0.1:9080",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm.startMock() // Mockにコンテキストが欲しい
		time.Sleep(10 * time.Second)
	})

	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("single worker with go-routine", func() {
		var machinesList Machines
		var err error
		var rslt, rslt2 SystemEventLog
		var serial1 string = "683FPQ3"
		var file *os.File
		var reader *bufio.Reader

		It("get machine list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("Machine List = ", machinesList)
		})

		// ワーカースレッドの停止のテストが欲しい

		// Start log collector
		It("run worker with the go routine (1st time)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.worker(ctx, machinesList.Machine[i])
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(3 * time.Second)
			}
			defer cancel()
		})

		It("verify output of collector (1st time)", func() {
			for {
				file, err = os.Open(path.Join(lc.testOut, serial1))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				reader = bufio.NewReaderSize(file, 4096)
				stringJSON, _ := reader.ReadString('\n')
				json.Unmarshal([]byte(stringJSON), &rslt)
				GinkgoWriter.Println("------ ", string(rslt.Serial))
				GinkgoWriter.Println("------ ", string(rslt.Id))
				break
			}
			Expect(rslt.Serial).To(Equal(serial1))
			Expect(rslt.Id).To(Equal("1"))
		})

		// Start log collector
		It("run worker with the go routine (2nd time)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			GinkgoWriter.Println("------ ", machinesList.Machine)
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.worker(ctx, machinesList.Machine[i])
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(3 * time.Second)
			}
			defer cancel()
		})

		It("verify output of collector (2nd time)", func() {
			stringJSON, _ := reader.ReadString('\n')
			fmt.Println("*3 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial1))
			Expect(rslt2.Id).To(Equal("2"))
		})
		file.Close()

	})
	AfterAll(func() {
		fmt.Println("shutdown workers")
		time.Sleep(5 * time.Second)
	})
})
