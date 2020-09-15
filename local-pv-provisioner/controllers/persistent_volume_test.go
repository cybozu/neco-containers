package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type deleterMock struct {
}

func (deleterMock) Delete(path string) error {
	return nil
}

func testPersistentVolumeReconciler() {
	It("normal case", func() {

	})
}
