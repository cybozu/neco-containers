package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	var mq MessageQueue

	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		os.Remove("testdata/pointers/HN3CLP3")
		os.Remove("testdata/output/HN3CLP3")

		os.Remove("testdata/pointers/J7N6MW3")
		os.Remove("testdata/output/J7N6MW3")

		ctx, cancel := context.WithCancel(context.Background())
		mq.queue = make(chan Machine, 1000)

		lc = logCollector{
			machinesPath: "testdata/configmap/serverlist2.csv",
			miniNum:      1,  // 最小
			maxiNum:      10, // 最大
			currNum:      1,  // 決定コレクター数
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			ctx:          ctx, // コンテキスト
			cancel:       cancel,
			que:          mq, // コレクターのキュー
			interval:     3,  // 待機秒数
			wg:           &wg,
			testMode:     true,
			testOut:      "testdata/output",
		}
		//defer cancel()
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
		bm3.startMock()

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

		var rslt1, rslt2, rslt3 SystemEventLog
		var serial1 string = "683FPQ3"
		var serial2 string = "HN3CLP3"
		var serial3 string = "J7N6MW3"
		var file1, file2, file3 *os.File
		var reader1, reader2, reader3 *bufio.Reader

		It("read CSV file", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		// ３サイクル分の処理をキューに積む
		It("put que 1", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put3(machinesList.machine)
			fmt.Println("que len = ", lc.que.len2())
		}, SpecTimeout(3*time.Second))

		It("put que 2", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put3(machinesList.machine)
			fmt.Println("que len = ", lc.que.len2())
		}, SpecTimeout(3*time.Second))

		It("put que 3", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put3(machinesList.machine)
			fmt.Println("que len = ", lc.que.len2())
		}, SpecTimeout(3*time.Second))

		// 複数のワーカーを起動
		It("start three worker", func() {
			for i := 0; i < lc.currNum; i++ {
				go lc.worker(i)
			}
			// 起動待ち
			time.Sleep(10 * time.Second)
		})

		It("Check 1st reply from iDRAC #1", func() {
			// ファイルが生成されるまで待つ
			for {
				file1, err = os.Open(path.Join(lc.testOut, serial1))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}

			reader1 = bufio.NewReaderSize(file1, 4096)
			stringJSON, _ := reader1.ReadString('\n')
			GinkgoWriter.Println("**** 11 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt1)
			GinkgoWriter.Println("------ ", string(rslt1.Serial))
			GinkgoWriter.Println("------ ", string(rslt1.Id))
			Expect(rslt1.Serial).To(Equal(serial1))
			Expect(rslt1.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #1", func() {
			rslt1 = SystemEventLog{}
			var stringJSON string

			// ファイルが生成されるまで待つ
			for {
				stringJSON, err = reader1.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			fmt.Println("**** 12 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt1)
			GinkgoWriter.Println("------ ", string(rslt1.Serial))
			GinkgoWriter.Println("------ ", string(rslt1.Id))

			Expect(rslt1.Serial).To(Equal(serial1))
			Expect(rslt1.Id).To(Equal("2"))
		})

		It("Check 1st reply from iDRAC #2", func() {
			for {
				file2, err = os.Open(path.Join(lc.testOut, serial2))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}
			reader2 = bufio.NewReaderSize(file2, 4096)
			stringJSON, _ := reader2.ReadString('\n')

			GinkgoWriter.Println("*1 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))

			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #2", func() {
			var stringJSON string
			rslt2 = SystemEventLog{}

			// ファイルが追記されるまで待機
			for {
				stringJSON, err = reader2.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*2 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("2"))
		})

		It("Check 3rd reply from iDRAC #2", func() {
			var stringJSON string
			rslt2 = SystemEventLog{}
			for {
				stringJSON, err = reader2.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*3 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("3"))
		})

		It("Check 4th reply from iDRAC #2", func() {
			var stringJSON string
			rslt2 = SystemEventLog{}

			for {
				stringJSON, err = reader2.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*4 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("4"))
		})

		It("Check 1st reply from iDRAC #3", func() {
			for {
				file3, err = os.Open(path.Join(lc.testOut, serial3))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}
			reader3 = bufio.NewReaderSize(file3, 4096)
			stringJSON, _ := reader3.ReadString('\n')
			GinkgoWriter.Println("*1 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt3)
			GinkgoWriter.Println("------ ", string(rslt3.Serial))
			GinkgoWriter.Println("------ ", string(rslt3.Id))

			Expect(rslt3.Serial).To(Equal(serial3))
			Expect(rslt3.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #3", func() {
			rslt3 = SystemEventLog{}
			var stringJSON string

			for {
				stringJSON, err = reader3.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*2 stringJSON=", stringJSON)

			json.Unmarshal([]byte(stringJSON), &rslt3)
			GinkgoWriter.Println("------ ", string(rslt3.Serial))
			GinkgoWriter.Println("------ ", string(rslt3.Id))

			Expect(rslt3.Serial).To(Equal(serial3))
			Expect(rslt3.Id).To(Equal("2"))
		})

		It("Check 3rd reply from iDRAC #3 after SEL cleanup", func() {
			rslt3 = SystemEventLog{}
			var stringJSON string

			for {
				stringJSON, err = reader3.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*3 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt3)

			GinkgoWriter.Println("------ ", string(rslt3.Serial))
			GinkgoWriter.Println("------ ", string(rslt3.Id))
			Expect(rslt3.Serial).To(Equal(serial3))
			Expect(rslt3.Id).To(Equal("1"))
		})

		It("Check 4th reply from iDRAC #3 after SEL cleanup", func() {
			rslt3 = SystemEventLog{}
			var stringJSON string

			for {
				stringJSON, err = reader3.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*3 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt3)

			GinkgoWriter.Println("------ ", string(rslt3.Serial))
			GinkgoWriter.Println("------ ", string(rslt3.Id))
			Expect(rslt3.Serial).To(Equal(serial3))
			Expect(rslt3.Id).To(Equal("2"))
		})

		It("Check 5th reply from iDRAC #3 after SEL cleanup", func() {
			rslt3 = SystemEventLog{}
			var stringJSON string

			for {
				stringJSON, err = reader3.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				}
				break
			}
			GinkgoWriter.Println("*3 stringJSON=", stringJSON)
			json.Unmarshal([]byte(stringJSON), &rslt3)

			GinkgoWriter.Println("------ ", string(rslt3.Serial))
			GinkgoWriter.Println("------ ", string(rslt3.Id))
			Expect(rslt3.Serial).To(Equal(serial3))
			Expect(rslt3.Id).To(Equal("3"))
		})

	})

	AfterAll(func() {
		fmt.Println("shutdown workers")
		lc.cancel()
		time.Sleep(5 * time.Second)
	})
})
