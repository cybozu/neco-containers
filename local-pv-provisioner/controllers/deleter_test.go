package controllers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/google/go-cmp/cmp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testFillDeleter() {
	It("should fill first specified bytes with zero", func() {
		tmpFile, _ := ioutil.TempFile("", "deleter")
		defer os.Remove(tmpFile.Name())
		err := exec.Command("dd", `if=/dev/urandom`, "of="+tmpFile.Name(), fmt.Sprintf("bs=%d", 1024), "count=11").Run()
		Expect(err).ShouldNot(HaveOccurred())

		deleter := &FillDeleter{
			FillBlockSize: 1024,
			FillCount:     10,
		}
		deleter.Delete(tmpFile.Name())

		zeroBlock := make([]byte, deleter.FillBlockSize)
		buffer := make([]byte, deleter.FillBlockSize)
		for i := uint(0); i < deleter.FillCount; i++ {
			_, err := tmpFile.Read(buffer)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cmp.Equal(buffer, zeroBlock)).Should(BeTrue())
		}

		_, err = tmpFile.Read(buffer)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cmp.Equal(buffer, zeroBlock)).Should(BeFalse())
	})
}
