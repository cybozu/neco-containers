package main

/*
  - Read the machine list and access iDRAC mock.
  - Verify to function of parallel collection logs from iDRAC mock
  - Verify anti-duplicate filter.
  - Verify identify the latest record when iDRAC log clear.
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting by parallel workers", Ordered, func() {

	var wg sync.WaitGroup
	var lc logCollector

	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		os.Remove("testdata/pointers/HN3CLP3")
		os.Remove("testdata/output/HN3CLP3")

		os.Remove("testdata/pointers/J7N6MW3")
		os.Remove("testdata/output/J7N6MW3")

		lc = logCollector{
			machinesPath: "testdata/configmap/serverlist2.json",
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			wg:           &wg,
			testMode:     true,
			testOut:      "testdata/output",
		}
		GinkgoWriter.Println("Start iDRAC Stub")

		bm1 := bmcMock{
			host:   "127.0.0.1:7180",
			resDir: "testdata/redfish_response",
			files:  []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"},
		}
		bm1.startMock()

		bm2 := bmcMock{
			host:   "127.0.0.1:7280",
			resDir: "testdata/redfish_response",
			files:  []string{"HN3CLP3-1.json", "HN3CLP3-2.json", "HN3CLP3-3.json"},
		}
		bm2.startMock()

		bm3 := bmcMock{
			host:   "127.0.0.1:7380",
			resDir: "testdata/redfish_response",
			files:  []string{"J7N6MW3-1.json", "J7N6MW3-2.json", "J7N6MW3-3.json"},
		}
		bm3.startMock() // コンテキストを入れる

		// Wait starting stub servers
		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("three workers", func() {
		var machinesList Machines
		var err error

		var serial1 string = "683FPQ3"
		var serial2 string = "HN3CLP3"
		var serial3 string = "J7N6MW3"
		var file1, file2, file3 *os.File
		var reader1, reader2, reader3 *bufio.Reader

		It("get target machines list", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		// Start Log collector in parallel.  Cycle=1
		It("run worker with the go routine (Cycle=1)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.worker(ctx, machinesList.Machine[i])
				//Expect(err).NotTo(HaveOccurred())
				time.Sleep(3 * time.Second)
			}
			defer cancel()
		})

		// Start Log collector in parallel.  Cycle=2
		It("run worker with the go routine (Cycle=2)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.worker(ctx, machinesList.Machine[i])
				//Expect(err).NotTo(HaveOccurred())
				time.Sleep(3 * time.Second)
			}
			defer cancel()
		})

		// Start Log collector in parallel.  Cycle=3
		It("run worker with the go routine (Cycle=3)", func() {
			ctx, cancel := context.WithCancel(context.Background())
			for i := 0; i < len(machinesList.Machine); i++ {
				go lc.worker(ctx, machinesList.Machine[i])
				//Expect(err).NotTo(HaveOccurred())
				time.Sleep(3 * time.Second)
			}
			defer cancel()
		})

		//////////////////////////////////////////////////////////////////////
		// Verify output for iDRAC #1 (serial1)
		It("verify 1st reply from iDRAC #1", func() {
			var rslt SystemEventLog
			file1, err = OpenTestResultLog(path.Join(lc.testOut, serial1))
			Expect(err).ToNot(HaveOccurred())

			// Read test log
			reader1 = bufio.NewReaderSize(file1, 4096)

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader1)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial1))
			Expect(rslt.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #1", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader1)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("---- serial = ", string(rslt.Serial))
			GinkgoWriter.Println("-------- id = ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial1))
			Expect(rslt.Id).To(Equal("2"))
		})

		//////////////////////////////////////////////////////////////////////
		// Verify output for iDRAC #2 (serial2)
		It("Check 1st reply from iDRAC #2", func() {
			var rslt SystemEventLog

			file2, err = OpenTestResultLog(path.Join(lc.testOut, serial2))
			Expect(err).ToNot(HaveOccurred())

			// Read test log
			reader2 = bufio.NewReaderSize(file2, 4096)

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader2)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial2))
			Expect(rslt.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #2", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader2)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial2))
			Expect(rslt.Id).To(Equal("2"))
		})

		It("Check 3rd reply from iDRAC #2", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader2)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial2))
			Expect(rslt.Id).To(Equal("3"))
		})

		It("Check 4th reply from iDRAC #2", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader2)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial2))
			Expect(rslt.Id).To(Equal("4"))
		})

		//////////////////////////////////////////////////////////////////////
		// Verify output for iDRAC #3 (serial3)
		It("Check 1st reply from iDRAC #3", func() {
			var rslt SystemEventLog

			file3, err = OpenTestResultLog(path.Join(lc.testOut, serial3))
			Expect(err).ToNot(HaveOccurred())

			// Read test log
			reader3 = bufio.NewReaderSize(file3, 4096)

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader3)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial3))
			Expect(rslt.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #3", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader3)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial3))
			Expect(rslt.Id).To(Equal("2"))
		})

		It("Check 3rd reply from iDRAC #3 after SEL cleanup", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader3)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial3))
			Expect(rslt.Id).To(Equal("1"))
		})

		It("Check 4th reply from iDRAC #3 after SEL cleanup", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader3)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial3))
			Expect(rslt.Id).To(Equal("2"))
		})

		It("Check 5th reply from iDRAC #3 after SEL cleanup", func() {
			var rslt SystemEventLog

			// Read test log
			stringJSON, err := ReadingTestResultLogNext(reader3)
			Expect(err).ToNot(HaveOccurred())
			GinkgoWriter.Println("**** Received stringJSON=", stringJSON)

			// JSON to struct
			err = json.Unmarshal([]byte(stringJSON), &rslt)
			Expect(err).ToNot(HaveOccurred())

			// Verify serial & id
			GinkgoWriter.Println("------ ", string(rslt.Serial))
			GinkgoWriter.Println("------ ", string(rslt.Id))
			Expect(rslt.Serial).To(Equal(serial3))
			Expect(rslt.Id).To(Equal("3"))
		})
	})

	AfterAll(func() {
		fmt.Println("shutdown workers")
		time.Sleep(5 * time.Second)
	})
})
