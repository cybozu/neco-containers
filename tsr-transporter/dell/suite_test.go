package dell_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTsrRequester(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TsrRequester Suite")
}
