package usecase

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/application/port/reader"
)

type systemReaderStub struct {
	metrics []reader.RuntimeMetric
	err     error
}

func (s *systemReaderStub) Read() ([]reader.RuntimeMetric, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.metrics, nil
}

func TestCollectSystemMetricsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name            string
		systemReader    *systemReaderStub
		repository      *metricRepositorySpy
		wantGauges      map[string]float64
		wantCountersLen int
		wantErrContains string
	}{
		{
			name: "stores system gauges",
			systemReader: &systemReaderStub{
				metrics: []reader.RuntimeMetric{
					{Name: "TotalMemory", Value: 1024},
					{Name: "FreeMemory", Value: 256},
					{Name: "CPUutilization1", Value: 10.5},
					{Name: "CPUutilization2", Value: 20.25},
				},
			},
			repository: &metricRepositorySpy{},
			wantGauges: map[string]float64{
				"TotalMemory":     1024,
				"FreeMemory":      256,
				"CPUutilization1": 10.5,
				"CPUutilization2": 20.25,
			},
		},
		{
			name: "normalizes invalid gauge values",
			systemReader: &systemReaderStub{
				metrics: []reader.RuntimeMetric{
					{Name: "TotalMemory", Value: math.NaN()},
					{Name: "FreeMemory", Value: math.Inf(1)},
				},
			},
			repository: &metricRepositorySpy{},
			wantGauges: map[string]float64{
				"TotalMemory": 0,
				"FreeMemory":  0,
			},
		},
		{
			name: "returns reader error",
			systemReader: &systemReaderStub{
				err: errors.New("system reader failed"),
			},
			repository:      &metricRepositorySpy{},
			wantErrContains: "system reader failed",
		},
		{
			name: "returns invalid metric name error",
			systemReader: &systemReaderStub{
				metrics: []reader.RuntimeMetric{
					{Name: "", Value: 1},
				},
			},
			repository:      &metricRepositorySpy{},
			wantErrContains: "name is empty",
		},
		{
			name: "returns repository error",
			systemReader: &systemReaderStub{
				metrics: []reader.RuntimeMetric{
					{Name: "TotalMemory", Value: 1024},
				},
			},
			repository: &metricRepositorySpy{
				setGaugeErr: errors.New("set gauge failed"),
			},
			wantErrContains: "set gauge failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewCollectSystemMetricsUseCase(tt.systemReader, tt.repository)

			// Act
			err := uc.Execute()

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

			gotGauges := gaugesToMap(tt.repository.gauges)
			if len(gotGauges) != len(tt.wantGauges) {
				t.Fatalf("expected %d gauges, got %d", len(tt.wantGauges), len(gotGauges))
			}

			for name, wantValue := range tt.wantGauges {
				gotValue, ok := gotGauges[name]
				if !ok {
					t.Fatalf("expected gauge %q to be stored", name)
				}

				if gotValue != wantValue {
					t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}

			if len(tt.repository.counters) != tt.wantCountersLen {
				t.Fatalf("expected %d counters, got %d", tt.wantCountersLen, len(tt.repository.counters))
			}
		})
	}
}
