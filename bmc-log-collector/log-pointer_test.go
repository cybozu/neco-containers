package main

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"time"
)

var _ = Describe("Get Machines List", Ordered, func() {
	var ptr LastPointer
	var err error

	BeforeAll(func() {
		os.Mkdir("testdata/pointers", 0766)
		os.Remove("testdata/pointers/ABCDEF")
		file, _ := os.Create("testdata/pointers/WITHDRAWED")
		lptr := LastPointer{
			Serial:         "WITHDRAWED",
			LastReadTime:   0,
			LastReadId:     0,
			LastUpdateTime: time.Now().Unix() - (3600 * 24 * 30 * 6),
		}
		byteJSON, _ := json.Marshal(lptr)
		file.WriteString(string(byteJSON))
		file.Close()
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
