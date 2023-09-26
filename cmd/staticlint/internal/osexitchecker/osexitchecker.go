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
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.FuncDecl:
				if x.Name.String() != "main" {
					return false
				}
			case *ast.CallExpr:
				if selexpr, ok := x.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selexpr.X.(*ast.Ident); ok {
						if ident.Name == "os" && selexpr.Sel.Name == "Exit" {
							pass.Reportf(selexpr.Pos(), "calling os.Exit in main package main func")
						}
					}
				}
			}

			return true
		})
	}

	//nolint: nilnil // expected
	return nil, nil
}
