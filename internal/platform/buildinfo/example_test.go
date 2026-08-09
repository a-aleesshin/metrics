package buildinfo_test

import "github.com/a-aleesshin/metrics/internal/platform/buildinfo"

// ExamplePrint показывает формат вывода, включая замену пустых значений на N/A.
func ExamplePrint() {
	buildinfo.Print("v1.0.0", "", "5ddaff9")
	// Output:
	// Build version: v1.0.0
	// Build date: N/A
	// Build commit: 5ddaff9
}
