package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get Machines List", Ordered, func() {
	var ptr LastPointer
	var err error
	var testPointerDir = "testdata/pointers_get_machines"
	var serialNormal = "ABCDEF"
	var nodeIPNormal = "10.0.0.1"
	var serialForDelete = "WITHDRAWED"
	var nodeIPForDelete = "10.0.0.2"
	var ml []Machine

	BeforeAll(func() {
		os.Mkdir(testPointerDir, 0766)
		os.Remove(path.Join(testPointerDir, serialNormal))
		os.Remove(path.Join(testPointerDir, serialForDelete))

		// create pointer file for delete test
		file, _ := os.Create(path.Join(testPointerDir, serialForDelete))
		lptr := LastPointer{
			Serial:       serialForDelete,
			NodeIP:       nodeIPForDelete,
			LastReadTime: 0,
			LastReadId:   0,
		}
		byteJSON, _ := json.Marshal(lptr)
		file.WriteString(string(byteJSON))
		file.Close()
		// Set timestamps for past dates
		//pastTime := time.Now().UTC().AddDate(0, -6, 0).Format("200601021504.05")
		//exec.Command("touch", "-t", pastTime, path.Join(testPointerDir, serialForDelete)).Run()

		// create machines list for delete test
		m0 := Machine{
			Serial: "ABCDEF",
			BmcIP:  "10.0.0.1",
			NodeIP: "10.1.0.1",
			Role:   "cs",
			State:  "HEALTHY",
		}
		ml = append(ml, m0)

	})

	Context("normal JSON file", func() {
		It("read ptr file", func() {
			ptr, err = readLastPointer(serialNormal, nodeIPNormal, testPointerDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.Serial).To(Equal(serialNormal))
			Expect(ptr.NodeIP).To(Equal(nodeIPNormal))
			Expect(ptr.LastReadTime).To(Equal(int64(0)))
			Expect(ptr.LastReadId).To(Equal(0))
			GinkgoWriter.Println(ptr)
		})
		It("update ptr", func() {
			ptr.LastReadTime = 1
			ptr.LastReadId = 1
			err := updateLastPointer(ptr, testPointerDir)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	/*
		Context("Get machine list from pointer files", func() {
			It("get machine list", func() {
				m, err := getMachineListWhichEverAccessed(testPointerDir)
				Expect(err).NotTo(HaveOccurred())
				fmt.Println("machine list =", m)
				for k, v := range m {
					switch k {
					case "ABCDEF":
						Expect(v.Serial).To(Equal("ABCDEF"))
						Expect(v.NodeIP).To(Equal("10.0.0.1"))
					case "WITHDRAWED":
						Expect(v.Serial).To(Equal(serialForDelete))
						Expect(v.NodeIP).To(Equal(nodeIPForDelete))
					}
				}
			})
		})
	*/

	Context("delete retired server ptr file", func() {
		It("do delete", func() {
			fmt.Println("ML=", ml)
			err := deleteUnUpdatedFiles(testPointerDir, ml)
			Expect(err).NotTo(HaveOccurred())
		})
		It("check that file has been deleted", func() {
			filePath := path.Join(testPointerDir, serialForDelete)
			_, err := os.Open(filePath)
			Expect(err).To(HaveOccurred())
		})
	})

})
