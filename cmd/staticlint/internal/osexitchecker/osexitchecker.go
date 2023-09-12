package osexitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "checks of calling os.Exit in main package main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				if x, ok := node.(*ast.CallExpr); ok {
					selexpr, ok := x.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					ident, ok := selexpr.X.(*ast.Ident)
					if !ok || ident.Name != "os" {
						return true
					}
					if selexpr.Sel.Name == "Exit" {
						pass.Reportf(selexpr.Pos(), "calling os.Exit in main package main func")
					}
				}
				return true
			})
		}
	}

	//nolint: nilnil // expected
	return nil, nil
}
