package eventuallycheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer checks if you are forget to execute gomega.Assertion for gomega.Eventually
var Analyzer = &analysis.Analyzer{
	Name: "eventuallycheck",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	RunDespiteErrors: false,
}

const doc = "eventuallycheck checks if you forget to call Assertion for Eventually or not"

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

func isNamespacedEventuallyFunc(n ast.Expr, pkgName string) bool {
	switch x := n.(type) {
	case *ast.SelectorExpr:
		if !isIdent(x.X, pkgName) {
			return false
		}
		return x.Sel != nil && isEventuallyFunc(x.Sel)
	}
	return false
}

func isEventuallyCall(caller ast.Expr) bool {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isEventuallyFunc(x.Fun)
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		// for _, d := range f.Decls {
		// 	ast.Print(pass.Fset, d)
		// 	fmt.Println()
		// }
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
			// dot import
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.SelectorExpr:
					if isEventuallyCall(x.X) {
						// Eventually(<func definition>).Should(Succeed())
						return false
					}
				case *ast.CallExpr:
					if isEventuallyFunc(x.Fun) {
						// Eventually(<func definition>)
						pass.Reportf(n.Pos(), "invalid Eventually: Assertion not called")
						return false
					}
				}
				return true
			})
		} else {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.SelectorExpr:
					switch cx := x.X.(type) {
					case *ast.CallExpr:
						if isNamespacedEventuallyFunc(cx.Fun, gomegaPkgName) {
							// <pkgName>.Eventually(<func definition>).Should(Succeed())
							return false
						}
					}
				case *ast.CallExpr:
					if isNamespacedEventuallyFunc(x.Fun, gomegaPkgName) {
						// <pkgName>.Eventually(<func definition>)
						pass.Reportf(n.Pos(), "invalid Eventually: Assertion not called")
						return false
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
