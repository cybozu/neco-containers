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

	// setup queue
	var m sync.Mutex
	var wg sync.WaitGroup
	var q []Machine = make([]Machine, 0)
	var lc logCollector

	BeforeAll(func() {
		os.Remove("testdata/pointers/683FPQ3")
		os.Remove("testdata/output/683FPQ3")

		os.Remove("testdata/pointers/HN3CLP3")
		os.Remove("testdata/output/HN3CLP3")

		//os.Remove("testdata/pointers/J7N6MW3")
		//os.Remove("testdata/output/J7N6MW3")

		ctx, cancel := context.WithCancel(context.Background())
		mq := Queue{
			queue: q,
			mu:    &m,
		}
		lc = logCollector{
			machinesPath: "testdata/configmap/serverlist.csv",
			miniNum:      1,  // 最小
			maxiNum:      10, // 最大
			currNum:      2,  // 決定コレクター数
			rfUrl:        "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
			ptrDir:       "testdata/pointers",
			ctx:          ctx, // コンテキスト
			cancel:       cancel,
			que:          mq, // コレクターのキュー
			interval:     20, // 待機秒数
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

		//bm3 := bmcMock{
		//	host:   "127.0.0.1:7380",
		//	resDir: "testdata/redfish_response",
		//	files:  []string{"J7N6MW3-1.json", "J7N6MW3-2.json", "J7N6MW3-3.json"},
		//}
		//bm3.startMock()

		time.Sleep(10 * time.Second)
	})
	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
	})

	Context("three workers", func() {
		var machinesList Machines
		var err error

		/*
			var rslt1, rslt2, rslt3 SystemEventLog
			var serial1 string = "683FPQ3"
			var serial2 string = "HN3CLP3"
			var serial3 string = "J7N6MW3"
			var file1, file2, file3 *os.File
			var reader1, reader2, reader3 *bufio.Reader
		*/

		var rslt1, rslt2 SystemEventLog
		var serial1 string = "683FPQ3"
		var serial2 string = "HN3CLP3"
		var file1, file2 *os.File
		var reader1, reader2 *bufio.Reader

		It("read CSV file", func() {
			machinesList, err = machineListReader(lc.machinesPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("put que", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put(machinesList.machine)
		}, SpecTimeout(time.Second))

		// 複数のワーカーを起動
		It("start three worker", func() {
			for i := 0; i < lc.currNum; i++ {
				go lc.worker(i)
			}
			// 起動待ち
			time.Sleep(20 * time.Second)
		})

		///////////////////////////////////////////////////////////
		It("put que", func(ctx SpecContext) {
			GinkgoWriter.Println(machinesList.machine)
			lc.que.put(machinesList.machine)
		}, SpecTimeout(time.Second))

		It("Check 1st reply from iDRAC #1", func() {
			for {
				file1, err = os.Open(path.Join(lc.testOut, serial1))
				if errors.Is(err, os.ErrNotExist) {
					time.Sleep(3 * time.Second)
					continue
				}
				reader1 = bufio.NewReaderSize(file1, 4096)
				stringJSON, _ := reader1.ReadString('\n')
				fmt.Println("**** 11 stringJSON=", stringJSON)
				json.Unmarshal([]byte(stringJSON), &rslt1)
				GinkgoWriter.Println("------ ", string(rslt1.Serial))
				GinkgoWriter.Println("------ ", string(rslt1.Id))
				break
			}
			Expect(rslt1.Serial).To(Equal(serial1))
			Expect(rslt1.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #1", func() {
			var stringJSON string
			for {
				stringJSON, err = reader1.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
			fmt.Println("**** 12 stringJSON=", stringJSON)
			rslt1 = SystemEventLog{}
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
				reader2 = bufio.NewReaderSize(file2, 4096)
				stringJSON, _ := reader2.ReadString('\n')
				GinkgoWriter.Println("*1 stringJSON=", stringJSON)
				json.Unmarshal([]byte(stringJSON), &rslt2)
				GinkgoWriter.Println("------ ", string(rslt2.Serial))
				GinkgoWriter.Println("------ ", string(rslt2.Id))
				break
			}
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("1"))
		})

		It("Check 2nd reply from iDRAC #2", func() {
			time.Sleep(3 * time.Second)
			stringJSON, _ := reader2.ReadString('\n')
			GinkgoWriter.Println("*2 stringJSON=", stringJSON)
			rslt2 = SystemEventLog{}
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("2"))
		})

		It("Check 3rd reply from iDRAC #2", func() {
			time.Sleep(3 * time.Second)
			stringJSON, _ := reader2.ReadString('\n')
			GinkgoWriter.Println("*3 stringJSON=", stringJSON)
			rslt2 = SystemEventLog{}
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("3"))
		})

		It("Check 4th reply from iDRAC #2", func() {
			var stringJSON string
			//time.Sleep(10 * time.Second)
			for {
				stringJSON, err = reader2.ReadString('\n')
				if err == io.EOF {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
			//stringJSON, _ := reader2.ReadString('\n')
			GinkgoWriter.Println("*4 stringJSON=", stringJSON)
			rslt2 = SystemEventLog{}
			json.Unmarshal([]byte(stringJSON), &rslt2)
			GinkgoWriter.Println("------ ", string(rslt2.Serial))
			GinkgoWriter.Println("------ ", string(rslt2.Id))
			Expect(rslt2.Serial).To(Equal(serial2))
			Expect(rslt2.Id).To(Equal("4"))
		})

		/*
			It("Check 1st reply from iDRAC #3", func() {
				for {
					file3, err = os.Open(path.Join(lc.testOut, serial3))
					if errors.Is(err, os.ErrNotExist) {
						time.Sleep(3 * time.Second)
						continue
					}
					reader3 = bufio.NewReaderSize(file3, 4096)
					stringJSON, _ := reader3.ReadString('\n')
					GinkgoWriter.Println("*1 stringJSON=", stringJSON)
					json.Unmarshal([]byte(stringJSON), &rslt3)
					GinkgoWriter.Println("------ ", string(rslt3.Serial))
					GinkgoWriter.Println("------ ", string(rslt3.Id))
					break
				}
				Expect(rslt3.Serial).To(Equal(serial3))
				Expect(rslt3.Id).To(Equal("1"))
			})

			It("Check 2nd reply from iDRAC #3", func() {
				time.Sleep(3 * time.Second)
				stringJSON, _ := reader3.ReadString('\n')
				GinkgoWriter.Println("*2 stringJSON=", stringJSON)
				rslt3 = SystemEventLog{}
				json.Unmarshal([]byte(stringJSON), &rslt3)
				GinkgoWriter.Println("------ ", string(rslt3.Serial))
				GinkgoWriter.Println("------ ", string(rslt3.Id))
				Expect(rslt3.Serial).To(Equal(serial3))
				Expect(rslt3.Id).To(Equal("2"))
			})

			It("Check 3rd reply from iDRAC #3", func() {
				time.Sleep(3 * time.Second)
				stringJSON, _ := reader3.ReadString('\n')
				GinkgoWriter.Println("*3 stringJSON=", stringJSON)
				rslt3 = SystemEventLog{}
				json.Unmarshal([]byte(stringJSON), &rslt3)
				GinkgoWriter.Println("------ ", string(rslt3.Serial))
				GinkgoWriter.Println("------ ", string(rslt3.Id))
				Expect(rslt3.Serial).To(Equal(serial3))
				Expect(rslt3.Id).To(Equal("1"))
			})

			It("Check 4th reply from iDRAC #3", func() {
				time.Sleep(3 * time.Second)
				stringJSON, _ := reader3.ReadString('\n')
				GinkgoWriter.Println("*4 stringJSON=", stringJSON)
				rslt3 = SystemEventLog{}
				json.Unmarshal([]byte(stringJSON), &rslt3)
				GinkgoWriter.Println("------ ", string(rslt3.Serial))
				GinkgoWriter.Println("------ ", string(rslt3.Id))
				Expect(rslt3.Serial).To(Equal(serial3))
				Expect(rslt3.Id).To(Equal("2"))
			})

			It("Check 5th reply from iDRAC #3", func() {
				var stringJSON string
				for {
					stringJSON, err = reader3.ReadString('\n')
					if err == io.EOF {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
				//stringJSON, err := reader3.ReadString('\n')
				//if err != nil {
				//	fmt.Println(err)
				//}
				GinkgoWriter.Println("*5 stringJSON=", stringJSON)
				rslt3 = SystemEventLog{}
				err = json.Unmarshal([]byte(stringJSON), &rslt3)
				if err != nil {
					fmt.Println(err)
				}
				GinkgoWriter.Println("------ ", string(rslt3.Serial))
				GinkgoWriter.Println("------ ", string(rslt3.Id))
				Expect(rslt3.Serial).To(Equal(serial3))
				Expect(rslt3.Id).To(Equal("3"))
			})

				It("Check 5th reply from iDRAC #3", func() {
					time.Sleep(20 * time.Second)
					stringJSON, _ := reader3.ReadString('\n')
					GinkgoWriter.Println("*5 stringJSON=", stringJSON)
					rslt3 = SystemEventLog{}
					json.Unmarshal([]byte(stringJSON), &rslt3)
					GinkgoWriter.Println("------ ", string(rslt3.Serial))
					GinkgoWriter.Println("------ ", string(rslt3.Id))
					Expect(rslt3.Serial).To(Equal(serial3))
					Expect(rslt3.Id).To(Equal("2"))
				})
		*/
	})

	AfterAll(func() {
		fmt.Println("shutdown workers")
		lc.cancel()
		time.Sleep(5 * time.Second)
	})

})
