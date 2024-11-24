package mainexitanalyzer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name:     "mainexitanalyzer",
	Doc:      "Checking of use call os.Exit() in main function",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		if pass.Pkg.Name() != "main" {
			return nil, nil
		}

		ast.Inspect(file, func(node ast.Node) bool {

			switch val := node.(type) {
			case *ast.FuncDecl:
				if val.Name.Name != "main" {
					return true
				}

				ast.Inspect(val, func(node ast.Node) bool {
					callExprNodeType, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}

					selExpr, ok := callExprNodeType.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					if idExpr, ok := selExpr.X.(*ast.Ident); ok && idExpr.Name == "os" && selExpr.Sel.Name == "Exit" {
						pass.Reportf(selExpr.Pos(), "allow not using call os.Exit() in main function")
					}
					return false
				})

			}
			return true
		})
	}

	return nil, nil
}
