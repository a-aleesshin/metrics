// Package nopanic содержит анализатор, сообщающий об использовании
// встроенной функции panic.
//
// panic обрывает нормальный поток управления и роняет процесс, если его
// не перехватить recover; в сервисном коде ошибки следует возвращать
// значениями. Анализатор различает встроенный panic и одноимённые
// пользовательские идентификаторы через информацию о типах.
package nopanic

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer сообщает о вызовах встроенной функции panic.
// Сгенерированные файлы пропускаются.
var Analyzer = &analysis.Analyzer{
	Name: "nopanic",
	Doc:  "сообщает об использовании встроенной функции panic",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := call.Fun.(*ast.Ident)
			if !ok || ident.Name != "panic" {
				return true
			}

			if _, ok := pass.TypesInfo.Uses[ident].(*types.Builtin); !ok {
				return true
			}

			pass.Reportf(call.Pos(), "обнаружен вызов встроенной функции panic")

			return true
		})
	}

	return nil, nil
}
