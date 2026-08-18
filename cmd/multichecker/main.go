package main

import (
	"log"
	"strings"

	"github.com/a-aleesshin/metrics/cmd/staticlint/nopanic"
	"github.com/a-aleesshin/metrics/cmd/staticlint/osexit"
	"github.com/gostaticanalysis/nilerr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	multichecker.Main(buildAnalyzers()...)
}

// buildAnalyzers собирает полный набор анализаторов multichecker:
// стандартные passes, все SA staticcheck, выбранные S/ST/QF,
// публичные bodyclose и nilerr, собственные osexit и nopanic.
func buildAnalyzers() []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	// Все анализаторы класса SA (staticcheck).
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// По одному анализатору остальных классов staticcheck.io.
	analyzers = append(analyzers,
		pick(simple.Analyzers, "S1005"),      // распаковка значения из map без лишней переменной
		pick(stylecheck.Analyzers, "ST1013"), // именованные константы вместо магических HTTP-кодов
		pick(quickfix.Analyzers, "QF1010"),   // байтовый срез в fmt: %s вместо string(...)
	)

	// Публичные анализаторы на выбор.
	analyzers = append(analyzers,
		bodyclose.Analyzer,
		nilerr.Analyzer,
	)

	// Собственные анализаторы.
	analyzers = append(analyzers, osexit.Analyzer, nopanic.Analyzer)

	return analyzers
}

// pick возвращает анализатор staticcheck по имени; при опечатке в имени
// завершает процесс, чтобы ошибка конфигурации обнаруживалась на старте.
func pick(analyzers []*lint.Analyzer, name string) *analysis.Analyzer {
	for _, v := range analyzers {
		if v.Analyzer.Name == name {
			return v.Analyzer
		}
	}

	log.Fatalf("multichecker: unknown analyzer %s", name)

	return nil
}
