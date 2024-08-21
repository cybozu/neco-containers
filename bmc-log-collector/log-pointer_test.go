package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get Machines List", Ordered, func() {
	var ptr LastPointer
	var err error
	var testPointerDir = "testdata/pointers_get_machines"
	var serial = "ABCDEF"
	var serialForDelete = "WITHDRAWED"

	BeforeAll(func() {
		os.Mkdir(testPointerDir, 0766)
		os.Remove(path.Join(testPointerDir, serial))
		file, _ := os.Create(path.Join(testPointerDir, serialForDelete))
		lptr := LastPointer{
			Serial:       serialForDelete,
			LastReadTime: 0,
			LastReadId:   0,
		}
		byteJSON, _ := json.Marshal(lptr)
		file.WriteString(string(byteJSON))
		file.Close()
		exec.Command("touch", "-t", "202401011200.00", path.Join(testPointerDir, serialForDelete)).Run()
	})

	Context("normal JSON file", func() {
		It("read ptr file", func() {
			ptr, err = readLastPointer(serial, testPointerDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.Serial).To(Equal(serial))
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

	Context("delete retired server ptr file", func() {
		It("do delete", func() {
			err := deleteUnUpdatedFiles(testPointerDir)
			Expect(err).NotTo(HaveOccurred())
		})
		It("check that file has been deleted", func() {
			filePath := path.Join(testPointerDir, serialForDelete)
			_, err := os.Open(filePath)
			Expect(err).To(HaveOccurred())
		})
	})

})
