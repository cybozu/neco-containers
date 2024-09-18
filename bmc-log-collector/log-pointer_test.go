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
		err := os.Mkdir(testPointerDir, 0766)
		Expect(err).NotTo(HaveOccurred())
		os.Remove(path.Join(testPointerDir, serialNormal))
		os.Remove(path.Join(testPointerDir, serialForDelete))

		// Create pointer file for delete test
		fd, _ := os.Create(path.Join(testPointerDir, serialForDelete))
		lptr := LastPointer{
			LastReadTime: 0,
			LastReadId:   0,
		}
		byteJSON, _ := json.Marshal(lptr)
		_, err = fd.WriteString(string(byteJSON))
		Expect(err).NotTo(HaveOccurred())
		fd.Close()

		//file, _ = os.Create(path.Join(testPointerDir, "_"+serialForDelete))
		//_, err = file.WriteString(string(byteJSON))
		//Expect(err).NotTo(HaveOccurred())
		//file.Close()

		// Create machines list for delete test
		m0 := Machine{
			Serial: "ABCDEF",
			BmcIP:  "10.0.0.1",
			NodeIP: "10.1.0.1",
		}
		ml = append(ml, m0)
	})

	Context("create the pointer file", func() {
		//var filePath string
		filePath := path.Join(testPointerDir, serialNormal)

		It("check and create pointer file", func() {
			err := checkAndCreatePointerFile(filePath)
			Expect(err).NotTo(HaveOccurred())
		})
		It("read ptr file", func() {
			ptr, err = readLastPointer(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LastReadTime).To(Equal(int64(0)))
			Expect(ptr.LastReadId).To(Equal(0))
			GinkgoWriter.Println(ptr)
		})
		It("update ptr", func() {
			ptr.LastReadTime = 1
			ptr.LastReadId = 1
			err := updateLastPointer(ptr, filePath)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("update the pointer file", func() {
		filePath := path.Join(testPointerDir, serialNormal)

		It("check and create pointer file", func() {
			err := checkAndCreatePointerFile(filePath)
			Expect(err).NotTo(HaveOccurred())
		})
		It("read ptr file", func() {
			ptr, err = readLastPointer(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptr.LastReadTime).To(Equal(int64(1)))
			Expect(ptr.LastReadId).To(Equal(1))
			GinkgoWriter.Println(ptr)
		})
		It("update ptr", func() {
			ptr.LastReadTime = 2
			ptr.LastReadId = 2
			err := updateLastPointer(ptr, filePath)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	/*
		Context("recover from the pointer file lost", func() {
			filePath := path.Join(testPointerDir, serialNormal)

			It("Remove the primary pointer file", func() {
				err := os.Remove(filePath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("check and create pointer file", func() {
				filePath = path.Join(testPointerDir, serialNormal)
				err := checkAndCreatePointerFile(filePath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("read ptr file", func() {
				ptr, err = readLastPointer(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(ptr.LastReadTime).To(Equal(int64(2)))
				Expect(ptr.LastReadId).To(Equal(2))
				GinkgoWriter.Println(ptr)
			})
			It("update ptr", func() {
				ptr.LastReadTime = 3
				ptr.LastReadId = 3
				err := updateLastPointer(ptr, filePath)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	*/

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
