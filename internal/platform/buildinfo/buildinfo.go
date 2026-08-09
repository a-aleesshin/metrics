// Package buildinfo печатает информацию о сборке бинарника: версию, дату
// и коммит
//
//	go build -ldflags "-X main.buildVersion=v1.0.0 \
//	  -X main.buildDate=2026-08-09 \
//	  -X main.buildCommit=5ddaff9" ./cmd/server
package buildinfo

import "fmt"

// Print выводит в stdout версию, дату и коммит сборки.
func Print(version, date, commit string) {
	fmt.Printf("Build version: %s\n", orNA(version))
	fmt.Printf("Build date: %s\n", orNA(date))
	fmt.Printf("Build commit: %s\n", orNA(commit))
}

func orNA(value string) string {
	if value == "" {
		return "N/A"
	}

	return value
}
