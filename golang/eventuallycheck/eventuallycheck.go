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

const errorMessage = "invalid Eventually: Assertion not called"

func isIdent(n ast.Expr, name string) bool {
	switch x := n.(type) {
	case *ast.Ident:
		if x.Name == name {
			return true
		}
	}
	return false
}

func isEventuallyFunc(n ast.Expr) bool {
	return isIdent(n, "Eventually")
}

func isNamedEventuallyFunc(n ast.Expr, pkgName string) bool {
	switch x := n.(type) {
	case *ast.SelectorExpr:
		if !isIdent(x.X, pkgName) {
			return false
		}
		return x.Sel != nil && isEventuallyFunc(x.Sel)
	}
	return false
}

func isEventuallyCalled(caller ast.Expr) bool {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isEventuallyFunc(x.Fun)
	}
	return false
}

func isNamedEventuallyCalled(caller ast.Expr, gomegaPkgName string) bool {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isNamedEventuallyFunc(x.Fun, gomegaPkgName)
	}
	return false
}

func checkForNamedImportFile(f *ast.File, gomegaPkgName string, pass *analysis.Pass) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// <pkgName>.Eventually(<func definition>).Should(Succeed())
			return !isNamedEventuallyCalled(x.X, gomegaPkgName)
		case *ast.CallExpr:
			if isNamedEventuallyFunc(x.Fun, gomegaPkgName) {
				// <pkgName>.Eventually(<func definition>)
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
			// Eventually(<func definition>).Should(Succeed())
			return !isEventuallyCalled(x.X)
		case *ast.CallExpr:
			if isEventuallyFunc(x.Fun) {
				// Eventually(<func definition>)
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
