package restrictpkg_test

import (
	"testing"

	"github.com/cybozu/neco-containers/golang/analyzer/pkg/restrictpkg"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := restrictpkg.RestrictPackageAnalyzer
	err := analyzer.Flags.Set("packages", "html/template")
	if err != nil {
		panic(err)
	}
	analysistest.Run(t, testdata, analyzer, "a", "b")
}
