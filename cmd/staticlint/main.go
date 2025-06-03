// Модуль main служит для реализации анализатора.
package main

import (
	"go/ast"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/ast/inspector"
	"honnef.co/go/tools/staticcheck"
)

// Main - функция для запуска анализатора.
func main() {
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)

	}

	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, shadow.Analyzer)
	mychecks = append(mychecks, structtag.Analyzer)
	mychecks = append(mychecks, noOsExitInMain)
	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, errcheck.Analyzer)

	multichecker.Main(
		mychecks...,
	)

}

var noOsExitInMain = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "prohibits direct calls to os.Exit in main.main function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтр — интересуют объявления функций
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)

		// Проверяем, что это функция main в пакете main
		if fn.Name.Name == "main" && pass.Pkg.Name() == "main" {
			// Обходим тело функции main
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Проверяем, что вызов — os.Exit
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Lparen, "direct call to os.Exit in main.main is forbidden")
				}

				return true
			})
		}
	})

	return nil, nil
}
