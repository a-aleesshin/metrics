package main

import (
	"github.com/a-aleesshin/metrics/cmd/staticlint/osexit"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(osexit.Analyzer)
}
