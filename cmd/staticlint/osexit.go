package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalizer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "restrict to call direct os.Exit in main",
	Run:  runOsExit,
}

func runOsExit(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Name.Name != "main" {
				return true
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				if callExpr, ok := n.(*ast.CallExpr); ok {
					if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := selector.X.(*ast.Ident); ok {
							if ident.Name == "os" && selector.Sel.Name == "Exit" {
								pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is prohibited")
							}
						}
					}
				}
				return true
			})

			return true
		})
	}

	return nil, nil
}
