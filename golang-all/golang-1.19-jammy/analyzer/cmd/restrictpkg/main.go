package main

import (
	"github.com/cybozu/neco-containers/golang/analyzer/pkg/restrictpkg"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(restrictpkg.RestrictPackageAnalyzer)
}
