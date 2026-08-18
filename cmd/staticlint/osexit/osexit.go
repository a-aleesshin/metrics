// Package osexit содержит анализатор, запрещающий прямой вызов os.Exit
// в функции main пакета main.
//
// Прямой os.Exit в main завершает процесс мгновенно: не выполняются defer,
// не сбрасываются буферы логов, не закрываются ресурсы. Правильный паттерн —
// вынести логику в функцию, возвращающую ошибку, и завершать процесс
// в одной точке (например, через log.Fatal).
package osexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer запрещает прямой вызов os.Exit в функции main пакета main.
// Сгенерированные файлы (в том числе тестовые main, создаваемые go test)
// пропускаются.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				if isOsExit(pass, call) {
					pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main пакета main запрещён")
				}

				return true
			})
		}
	}

	return nil, nil
}

// isOsExit проверяет через информацию о типах, что вызов — именно функция
// Exit пакета os, а не одноимённый метод или функция другого пакета.
func isOsExit(pass *analysis.Pass, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Exit" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return false
	}

	return pkgName.Imported().Path() == "os"
}
