package osexit_test

import (
	"testing"

	"github.com/a-aleesshin/metrics/cmd/staticlint/osexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexit.Analyzer, "a", "b")
}
