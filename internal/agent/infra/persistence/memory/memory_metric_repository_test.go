package memory

import (
	"testing"

	"github.com/a-aleesshin/metrics/internal/agent/domain/metric"
)

func TestMemMetricRepository(t *testing.T) {
	tests := []struct {
		name         string
		prepare      func(t *testing.T, repo *MemMetricRepository)
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			name: "set gauge stores value",
			prepare: func(t *testing.T, repo *MemMetricRepository) {
				metricName, err := metric.NewName("Alloc")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				err = repo.SetGauge(metric.NewGauge(metricName, 123.45))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			wantGauges: map[string]float64{
				"Alloc": 123.45,
			},
			wantCounters: map[string]int64{},
		},
		{
			name: "set gauge overwrites previous value",
			prepare: func(t *testing.T, repo *MemMetricRepository) {
				metricName, err := metric.NewName("Alloc")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.SetGauge(metric.NewGauge(metricName, 100)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.SetGauge(metric.NewGauge(metricName, 200)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			wantGauges: map[string]float64{
				"Alloc": 200,
			},
			wantCounters: map[string]int64{},
		},
		{
			name: "add counter accumulates value",
			prepare: func(t *testing.T, repo *MemMetricRepository) {
				metricName, err := metric.NewName("PollCount")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.AddCounter(metric.NewCounter(metricName, 1)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.AddCounter(metric.NewCounter(metricName, 2)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			wantGauges: map[string]float64{},
			wantCounters: map[string]int64{
				"PollCount": 3,
			},
		},
		{
			name: "get metrics returns gauges and counters",
			prepare: func(t *testing.T, repo *MemMetricRepository) {
				gaugeName, err := metric.NewName("Alloc")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				counterName, err := metric.NewName("PollCount")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.SetGauge(metric.NewGauge(gaugeName, 123.45)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := repo.AddCounter(metric.NewCounter(counterName, 10)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			wantGauges: map[string]float64{
				"Alloc": 123.45,
			},
			wantCounters: map[string]int64{
				"PollCount": 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := NewMemMetricRepository()
			tt.prepare(t, repo)

			// Act
			state, err := repo.GetMetrics()

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotGauges := gaugesToMap(state.Gauges)
			gotCounters := countersToMap(state.Counters)

			if len(gotGauges) != len(tt.wantGauges) {
				t.Fatalf("expected %d gauges, got %d", len(tt.wantGauges), len(gotGauges))
			}

			for name, wantValue := range tt.wantGauges {
				gotValue, ok := gotGauges[name]
				if !ok {
					t.Fatalf("expected gauge %q to exist", name)
				}

				if gotValue != wantValue {
					t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}

			if len(gotCounters) != len(tt.wantCounters) {
				t.Fatalf("expected %d counters, got %d", len(tt.wantCounters), len(gotCounters))
			}

			for name, wantValue := range tt.wantCounters {
				gotValue, ok := gotCounters[name]
				if !ok {
					t.Fatalf("expected counter %q to exist", name)
				}

				if gotValue != wantValue {
					t.Fatalf("counter %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}
		})
	}
}

func gaugesToMap(gauges []metric.Gauge) map[string]float64 {
	out := make(map[string]float64, len(gauges))
	for _, gauge := range gauges {
		out[gauge.Name().String()] = gauge.Value()
	}
	return out
}

func countersToMap(counters []metric.Counter) map[string]int64 {
	out := make(map[string]int64, len(counters))
	for _, counter := range counters {
		out[counter.Name().String()] = counter.Value()
	}
	return out
}
