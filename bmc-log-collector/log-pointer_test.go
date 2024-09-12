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
	var serialForDelete = "WITHDRAWED"
	var ml []Machine

	BeforeAll(func() {
		os.Mkdir(testPointerDir, 0766)
		os.Remove(path.Join(testPointerDir, serialNormal))
		os.Remove(path.Join(testPointerDir, serialForDelete))

		// create pointer file for delete test
		file, _ := os.Create(path.Join(testPointerDir, serialForDelete))
		lptr := LastPointer{
			LastReadTime: 0,
			LastReadId:   0,
		}
		byteJSON, _ := json.Marshal(lptr)
		file.WriteString(string(byteJSON))
		file.Close()

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
			ptr, err = readLastPointer(serialNormal, testPointerDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LastReadTime).To(Equal(int64(0)))
			Expect(ptr.LastReadId).To(Equal(0))
			GinkgoWriter.Println(ptr)
		})
		It("update ptr", func() {
			ptr.LastReadTime = 1
			ptr.LastReadId = 1
			err := updateLastPointer(ptr, testPointerDir, serialNormal)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("delete retired server ptr file", func() {
		It("do delete", func() {
			fmt.Println("ML=", ml)
			err := deletePtrFileDisappearedSerial(testPointerDir, ml)
			Expect(err).NotTo(HaveOccurred())
		})
		It("check that file has been deleted", func() {
			filePath := path.Join(testPointerDir, serialForDelete)
			_, err := os.Open(filePath)
			Expect(err).To(HaveOccurred())
		})
	})
})
