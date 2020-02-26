package eventuallycheck_test

import (
	"testing"

	"github.com/cybozu/neco-containers/golang/eventuallycheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := eventuallycheck.Analyzer
	analysistest.Run(t, testdata, analyzer, "dotimport", "import1", "import2", "override")
}
