package runtimeadapter

import "testing"

func TestMetricRuntimeReader_Read(t *testing.T) {
	tests := []struct {
		name      string
		wantNames []string
	}{
		{
			name: "returns all required runtime metrics",
			wantNames: []string{
				"Alloc",
				"BuckHashSys",
				"Frees",
				"GCCPUFraction",
				"GCSys",
				"HeapAlloc",
				"HeapIdle",
				"HeapInuse",
				"HeapObjects",
				"HeapReleased",
				"HeapSys",
				"LastGC",
				"Lookups",
				"MCacheInuse",
				"MCacheSys",
				"MSpanInuse",
				"MSpanSys",
				"Mallocs",
				"NextGC",
				"NumForcedGC",
				"NumGC",
				"OtherSys",
				"PauseTotalNs",
				"StackInuse",
				"StackSys",
				"Sys",
				"TotalAlloc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			reader := NewMetricRuntimeReader()

			// Act
			metrics := reader.Read()

			// Assert
			if len(metrics) != len(tt.wantNames) {
				t.Fatalf("expected %d metrics, got %d", len(tt.wantNames), len(metrics))
			}

			got := make(map[string]bool, len(metrics))
			for _, metric := range metrics {
				if metric.Name == "" {
					t.Fatal("metric name must not be empty")
				}

				if got[metric.Name] {
					t.Fatalf("duplicate metric name %q", metric.Name)
				}

				got[metric.Name] = true
			}

			for _, wantName := range tt.wantNames {
				if !got[wantName] {
					t.Fatalf("expected metric %q to be present", wantName)
				}
			}
		})
	}
}
