// Reset — генератор методов Reset() для структур, помеченных комментарием
// «// generate:reset».
//
// Утилита сканирует все пакеты модуля от корня и ниже, находит помеченные
// структуры и для каждого пакета записывает сгенерированные методы в файл
// reset.gen.go рядом с исходником. Повторный запуск перезаписывает файл.
//
// Запуск из корня репозитория:
//
//	go run ./cmd/reset
//
// Правила сброса полей:
//   - примитивы — к нулевым значениям (0, "", false);
//   - слайсы — обрезаются по длине (s = s[:0]), ёмкость сохраняется;
//   - мапы — очищаются встроенной clear;
//   - каналы, функции и интерфейсы — nil;
//   - вложенные структуры с методом Reset() — вызов Reset();
//   - структуры без Reset() и массивы — присваивание нулевого литерала;
//   - не-nil указатели — сброс значения по правилам выше.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

const (
	marker        = "generate:reset"
	generatedFile = "reset.gen.go"
)

// target — одна помеченная структура: имя, объект типа и каталог пакета.
type target struct {
	name string
	obj  *types.TypeName
	strc *types.Struct
	dir  string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	// Первый проход: собрать все цели по всем пакетам. Набор нужен заранее,
	// чтобы вызовы Reset() между целями работали и при первой генерации,
	// когда методов в загруженных типах ещё нет.
	perPkg := make(map[*packages.Package][]*target)
	targetSet := make(map[*types.TypeName]bool)

	for _, pkg := range pkgs {
		targets, err := findTargets(pkg)
		if err != nil {
			return err
		}

		if len(targets) == 0 {
			continue
		}

		perPkg[pkg] = targets

		for _, t := range targets {
			targetSet[t.obj] = true
		}
	}

	ordered := make([]*packages.Package, 0, len(perPkg))
	for pkg := range perPkg {
		ordered = append(ordered, pkg)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].PkgPath < ordered[j].PkgPath })

	for _, pkg := range ordered {
		targets := perPkg[pkg]

		src, err := generateFile(pkg, targets, targetSet)
		if err != nil {
			return fmt.Errorf("generate %s: %w", pkg.PkgPath, err)
		}

		path := filepath.Join(targets[0].dir, generatedFile)

		if err := os.WriteFile(path, src, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}

		log.Printf("generated %s (%d struct(s))", path, len(targets))
	}

	if len(ordered) == 0 {
		log.Printf("no structs marked with %q found", "// "+marker)
	}

	return nil
}

// findTargets возвращает помеченные структуры пакета в порядке имён.
func findTargets(pkg *packages.Package) ([]*target, error) {
	var targets []*target

	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		if filepath.Base(filename) == generatedFile {
			continue
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !hasMarker(gen.Doc, typeSpec.Doc, typeSpec.Comment) {
					continue
				}

				obj, ok := pkg.TypesInfo.Defs[typeSpec.Name].(*types.TypeName)
				if !ok {
					continue
				}

				strc, ok := obj.Type().Underlying().(*types.Struct)
				if !ok {
					return nil, fmt.Errorf("%s: %s помечен %q, но не является структурой",
						pkg.Fset.Position(typeSpec.Pos()), typeSpec.Name.Name, marker)
				}

				targets = append(targets, &target{
					name: typeSpec.Name.Name,
					obj:  obj,
					strc: strc,
					dir:  filepath.Dir(filename),
				})
			}
		}
	}

	sort.Slice(targets, func(i, j int) bool { return targets[i].name < targets[j].name })

	return targets, nil
}

// hasMarker проверяет, содержит ли хотя бы одна из групп комментариев
// строку-маркер generate:reset.
func hasMarker(groups ...*ast.CommentGroup) bool {
	for _, group := range groups {
		if group == nil {
			continue
		}

		for _, comment := range group.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text == marker {
				return true
			}
		}
	}

	return false
}

// generateFile собирает содержимое reset.gen.go для одного пакета
// и форматирует его go/format.
func generateFile(pkg *packages.Package, targets []*target, targetSet map[*types.TypeName]bool) ([]byte, error) {
	imports := make(map[string]bool)

	qualifier := func(p *types.Package) string {
		if p == pkg.Types {
			return ""
		}

		imports[p.Path()] = true

		return p.Name()
	}

	gen := &generator{
		qualifier: qualifier,
		targets:   targetSet,
		pkg:       pkg.Types,
	}

	var body bytes.Buffer
	for _, t := range targets {
		gen.writeReset(&body, t)
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "// Code generated by \"go run ./cmd/reset\". DO NOT EDIT.\n\n")
	fmt.Fprintf(&out, "package %s\n\n", pkg.Name)

	if len(imports) > 0 {
		paths := make([]string, 0, len(imports))
		for path := range imports {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		fmt.Fprintf(&out, "import (\n")
		for _, path := range paths {
			fmt.Fprintf(&out, "\t%q\n", path)
		}
		fmt.Fprintf(&out, ")\n\n")
	}

	out.Write(body.Bytes())

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w\n%s", err, out.String())
	}

	return formatted, nil
}

// generator пишет методы Reset() с учётом пакета назначения и набора целей.
type generator struct {
	qualifier types.Qualifier
	targets   map[*types.TypeName]bool
	pkg       *types.Package
}

func (g *generator) writeReset(w *bytes.Buffer, t *target) {
	recv := receiverName(t.name)

	fmt.Fprintf(w, "// Reset сбрасывает %s к начальным значениям.\n", t.name)
	fmt.Fprintf(w, "func (%s *%s) Reset() {\n", recv, t.name)
	fmt.Fprintf(w, "\tif %s == nil {\n\t\treturn\n\t}\n\n", recv)

	for i := 0; i < t.strc.NumFields(); i++ {
		field := t.strc.Field(i)
		g.writeFieldReset(w, recv+"."+field.Name(), field.Type(), 1)
	}

	fmt.Fprintf(w, "}\n\n")
}

// writeFieldReset пишет сброс одного выражения expr типа typ.
func (g *generator) writeFieldReset(w *bytes.Buffer, expr string, typ types.Type, depth int) {
	indent := strings.Repeat("\t", depth)

	if g.hasReset(typ) {
		fmt.Fprintf(w, "%s%s.Reset()\n", indent, expr)
		return
	}

	switch u := typ.Underlying().(type) {
	case *types.Basic:
		fmt.Fprintf(w, "%s%s = %s\n", indent, expr, zeroValue(u))

	case *types.Slice:
		fmt.Fprintf(w, "%s%s = %s[:0]\n", indent, expr, expr)

	case *types.Map:
		fmt.Fprintf(w, "%sclear(%s)\n", indent, expr)

	case *types.Chan, *types.Signature, *types.Interface:
		fmt.Fprintf(w, "%s%s = nil\n", indent, expr)

	case *types.Pointer:
		fmt.Fprintf(w, "%sif %s != nil {\n", indent, expr)
		g.writeFieldReset(w, "(*"+expr+")", u.Elem(), depth+1)
		fmt.Fprintf(w, "%s}\n", indent)

	case *types.Struct, *types.Array:
		fmt.Fprintf(w, "%s%s = %s{}\n", indent, expr, types.TypeString(typ, g.qualifier))

	default:
		fmt.Fprintf(w, "%s// %s: тип %s не поддерживается генератором\n", indent, expr, typ.String())
	}
}

// hasReset сообщает, нужно ли сбрасывать тип вызовом его метода Reset():
// тип либо сам помечен generate:reset, либо уже имеет метод Reset()
// без аргументов и результатов.
func (g *generator) hasReset(typ types.Type) bool {
	named, ok := types.Unalias(typ).(*types.Named)
	if !ok {
		return false
	}

	if g.targets[named.Obj()] {
		return true
	}

	if _, ok := named.Underlying().(*types.Struct); !ok {
		return false
	}

	obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(typ), true, g.pkg, "Reset")

	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}

	sig := fn.Type().(*types.Signature)

	return sig.Params().Len() == 0 && sig.Results().Len() == 0
}

// receiverName возвращает имя приёмника метода — первая буква имени типа
// в нижнем регистре.
func receiverName(typeName string) string {
	r, _ := utf8.DecodeRuneInString(typeName)
	if r == utf8.RuneError {
		return "x"
	}

	return string(unicode.ToLower(r))
}

// zeroValue возвращает выражение нулевого значения для примитивного типа.
func zeroValue(basic *types.Basic) string {
	info := basic.Info()

	switch {
	case info&types.IsBoolean != 0:
		return "false"
	case info&types.IsString != 0:
		return `""`
	case info&types.IsNumeric != 0:
		return "0"
	default:
		return "nil"
	}
}
