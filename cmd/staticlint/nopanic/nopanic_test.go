package nopanic_test

import (
	"testing"

	"github.com/a-aleesshin/metrics/cmd/staticlint/nopanic"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), nopanic.Analyzer, "a")
}
