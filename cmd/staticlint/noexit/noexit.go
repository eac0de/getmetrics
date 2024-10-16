package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer сообщает об использовании os.Exit в функции main.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "prohibits the use of os.Exit in the main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Ищем функцию main
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
							if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "os" && fun.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "ос.Exit should not be used in the main function")
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
