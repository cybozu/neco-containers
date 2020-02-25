package eventuallycheck

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// EventuallyCheckAnalyzer checks if you are forget to execute gomega.Eventually
var EventuallyCheckAnalyzer = &analysis.Analyzer{
	Name: "eventuallycheck",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	RunDespiteErrors: false,
}

const doc = "restrictpkg checks if you are forget to execute gomega.Eventually"

func isEventuallyFunc(fun ast.Expr) (bool, token.Pos) {
	switch x := fun.(type) {
	case *ast.Ident:
		if x.Name == "Eventually" {
			return true, x.NamePos
		}
	}
	return false, 0
}

func isEventuallyCall(caller ast.Expr) (bool, token.Pos) {
	switch x := caller.(type) {
	case *ast.CallExpr:
		return isEventuallyFunc(x.Fun)
	}
	return false, 0
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		// for _, d := range f.Decls {
		// 	ast.Print(pass.Fset, d)
		// 	fmt.Println() // \n したい...
		// }
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				b, pos := isEventuallyFunc(x.Fun)
				if b {
					fmt.Println("NG: " + pass.Fset.Position(pos).String())
				}
				return true
			case *ast.SelectorExpr:
				b, pos := isEventuallyCall(x.X)
				if b {
					fmt.Println("OK: " + pass.Fset.Position(pos).String())
				}
				return false
			}
			return true
		})
	}
	return nil, nil
}

// func main() {
// 	args := os.Args[1:]
// 	if len(args) == 0 {
// 		fmt.Println("need to specify target go file")
// 		os.Exit(1)
// 	}
// 	targetFile := args[0]

// 	fset := token.NewFileSet()
// 	f, err := parser.ParseFile(fset, targetFile, nil, parser.Mode(0))
// 	if err != nil {
// 		panic(err)
// 	}

// 	ast.Inspect(f, func(n ast.Node) bool {
// 		switch x := n.(type) {
// 		case *ast.CallExpr:
// 			b, pos := isEventuallyFunc(x.Fun)
// 			if b {
// 				fmt.Println("NG: " + fset.Position(pos).String())
// 			}
// 			return true
// 		case *ast.SelectorExpr:
// 			b, pos := isEventuallyCall(x.X)
// 			if b {
// 				fmt.Println("OK: " + fset.Position(pos).String())
// 			}
// 			return false
// 		}
// 		return true
// 	})
// }
