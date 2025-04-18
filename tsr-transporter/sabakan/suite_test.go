package sabakan_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSabakanApiLib(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sabakan API Suite")
}
