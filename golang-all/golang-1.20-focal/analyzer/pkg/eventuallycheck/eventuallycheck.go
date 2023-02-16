package eventuallycheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks if you forget to execute gomega.Assertion for gomega.Eventually
var Analyzer = &analysis.Analyzer{
	Name:             "eventuallycheck",
	Doc:              "eventuallycheck checks if you forget to call Assertion for Eventually or not",
	Run:              run,
	RunDespiteErrors: false,
}

var assertionFuncs = []string{
	"Consistently",
	"ConsistentlyWithOffset",
	"Eventually",
	"EventuallyWithOffset",
	"Expect",
	"ExpectWithOffset",
	"Ω",
}

const errorMessage = "invalid Assertion: Should/ShouldNot not called"

func isIdent(n ast.Expr, names ...string) bool {
	switch x := n.(type) {
	case *ast.Ident:
		for _, n := range names {
			if x.Name == n {
				return true
			}
		}
	}
	return false
}

func isAssertionFunc(n ast.Expr) bool {
	return isIdent(n, assertionFuncs...)
}

func isNamedAssertionFunc(n ast.Expr, pkgName string) bool {
	switch x := n.(type) {
	case *ast.SelectorExpr:
		if !isIdent(x.X, pkgName) {
			return false
		}
		return x.Sel != nil && isAssertionFunc(x.Sel)
	}
	return false
}

func isAssertionFuncCalled(caller ast.Expr) bool {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isAssertionFunc(x.Fun)
	}
	return false
}

func isNamedAssertionFuncCalled(caller ast.Expr, gomegaPkgName string) bool {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isNamedAssertionFunc(x.Fun, gomegaPkgName)
	}
	return false
}

func checkForNamedImportFile(f *ast.File, gomegaPkgName string, pass *analysis.Pass) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// e.g. <pkgName>.Eventually(<func definition>).Should(Succeed())
			return !isNamedAssertionFuncCalled(x.X, gomegaPkgName)
		case *ast.CallExpr:
			if isNamedAssertionFunc(x.Fun, gomegaPkgName) {
				// e.g. <pkgName>.Eventually(<func definition>)
				pass.Reportf(n.Pos(), errorMessage)
				return false
			}
		}
		return true
	})
}

func checkForDotImportFile(f *ast.File, pass *analysis.Pass) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// e.g. Eventually(<func definition>).Should(Succeed())
			return !isAssertionFuncCalled(x.X)
		case *ast.CallExpr:
			if isAssertionFunc(x.Fun) {
				// e.g. Eventually(<func definition>)
				pass.Reportf(n.Pos(), errorMessage)
				return false
			}
		}
		return true
	})
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		useGomega := false
		gomegaPkgName := "gomega"

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ImportSpec:
				if x.Path == nil || x.Path.Value != `"github.com/onsi/gomega"` {
					return true
				}
				useGomega = true
				if x.Name != nil {
					gomegaPkgName = x.Name.Name
				}
			}
			return true
		})

		if !useGomega {
			continue
		}

		if gomegaPkgName == "." {
			checkForDotImportFile(f, pass)
		} else {
			checkForNamedImportFile(f, gomegaPkgName, pass)
		}
	}
	return nil, nil
}
