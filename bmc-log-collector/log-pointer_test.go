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

	BeforeAll(func() {
		os.Mkdir("testdata/pointers", 0766)
		os.Remove("testdata/pointers/ABCDEF")
		file, _ := os.Create("testdata/pointers/WITHDRAWED")
		lptr := LastPointer{
			Serial:       "WITHDRAWED",
			LastReadTime: 0,
			LastReadId:   0,
			// 過去の成功と失敗を記録、連続２回の成功で、過去の失敗をリセットなど
			//LastUpdateTime: time.Now().Unix() - (3600 * 24 * 30 * 6),
		}
		byteJSON, _ := json.Marshal(lptr)
		file.WriteString(string(byteJSON))
		file.Close()
		exec.Command("touch", "-t", "202401011200.00", "testdata/pointers/WITHDRAWED").Run()
	})

	Context("normal JSON file", func() {
		It("read ptr file", func() {
			ptr, err = readLastPointer("ABCDEF", "testdata/pointers")
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.Serial).To(Equal("ABCDEF"))
			Expect(ptr.LastReadTime).To(Equal(int64(0)))
			Expect(ptr.LastReadId).To(Equal(0))
			GinkgoWriter.Println(ptr)
		})
		It("update ptr", func() {
			ptr.LastReadTime = 1
			ptr.LastReadId = 1
			err := updateLastPointer(ptr, "testdata/pointers")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("delete retired server ptr file", func() {
		It("do delete", func() {
			err := deleteUnUpdatedFiles("testdata/pointers")
			Expect(err).NotTo(HaveOccurred())
		})
		It("check that file has been deleted", func() {
			filePath := path.Join("testdata/pointers", "WITHDRAWED")
			_, err := os.Open(filePath)
			Expect(err).To(HaveOccurred())
		})
	})

	AfterAll(func() {
		os.Remove("testdata/pointers/ABCDEF")
	})
})
