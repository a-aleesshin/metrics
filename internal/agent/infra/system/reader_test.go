package systemadapter

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

func TestGopsutilReader_Read(t *testing.T) {
	tests := []struct {
		name              string
		virtualMemory     virtualMemoryReader
		cpuPercent        cpuPercentReader
		wantMetrics       map[string]float64
		wantErrContains   string
		wantCPUInterval   time.Duration
		wantCPUPerCPU     bool
		wantCPUWasInvoked bool
	}{
		{
			name: "returns memory and cpu metrics",
			virtualMemory: func() (*mem.VirtualMemoryStat, error) {
				return &mem.VirtualMemoryStat{
					Total: 1024,
					Free:  256,
				}, nil
			},
			cpuPercent: func(interval time.Duration, percpu bool) ([]float64, error) {
				return []float64{10.5, 20.25}, nil
			},
			wantMetrics: map[string]float64{
				"TotalMemory":     1024,
				"FreeMemory":      256,
				"CPUutilization1": 10.5,
				"CPUutilization2": 20.25,
			},
			wantCPUInterval:   0,
			wantCPUPerCPU:     true,
			wantCPUWasInvoked: true,
		},
		{
			name: "returns memory error",
			virtualMemory: func() (*mem.VirtualMemoryStat, error) {
				return nil, errors.New("memory failed")
			},
			cpuPercent: func(interval time.Duration, percpu bool) ([]float64, error) {
				return nil, nil
			},
			wantErrContains: "read virtual memory: memory failed",
		},
		{
			name: "returns cpu error",
			virtualMemory: func() (*mem.VirtualMemoryStat, error) {
				return &mem.VirtualMemoryStat{}, nil
			},
			cpuPercent: func(interval time.Duration, percpu bool) ([]float64, error) {
				return nil, errors.New("cpu failed")
			},
			wantErrContains:   "read cpu percent: cpu failed",
			wantCPUWasInvoked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var gotCPUInterval time.Duration
			var gotCPUPerCPU bool
			cpuWasInvoked := false

			reader := newGopsutilReader(
				tt.virtualMemory,
				func(interval time.Duration, percpu bool) ([]float64, error) {
					cpuWasInvoked = true
					gotCPUInterval = interval
					gotCPUPerCPU = percpu
					return tt.cpuPercent(interval, percpu)
				},
			)

			// Act
			metrics, err := reader.Read()

			// Assert
			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErrContains)
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotMetrics := make(map[string]float64, len(metrics))
			for _, metric := range metrics {
				gotMetrics[metric.Name] = metric.Value
			}

			if len(gotMetrics) != len(tt.wantMetrics) {
				t.Fatalf("expected %d metrics, got %d", len(tt.wantMetrics), len(gotMetrics))
			}

			for name, wantValue := range tt.wantMetrics {
				gotValue, ok := gotMetrics[name]
				if !ok {
					t.Fatalf("expected metric %q", name)
				}
				if gotValue != wantValue {
					t.Fatalf("metric %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}

			if cpuWasInvoked != tt.wantCPUWasInvoked {
				t.Fatalf("expected cpu invoked=%v, got %v", tt.wantCPUWasInvoked, cpuWasInvoked)
			}

			if cpuWasInvoked {
				if gotCPUInterval != tt.wantCPUInterval {
					t.Fatalf("expected cpu interval %v, got %v", tt.wantCPUInterval, gotCPUInterval)
				}
				if gotCPUPerCPU != tt.wantCPUPerCPU {
					t.Fatalf("expected percpu=%v, got %v", tt.wantCPUPerCPU, gotCPUPerCPU)
				}
			}
		})
	}
}
