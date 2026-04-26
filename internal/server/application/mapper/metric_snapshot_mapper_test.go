package mapper

import (
	"math"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func mustGauge(t *testing.T, id, name string, value float64) *metric.Gauge {
	t.Helper()

	gauge, err := metric.RestoreGauge(id, name, value)

	if err != nil {
		t.Fatalf("restore gauge: %v", err)
	}
	return gauge
}

func mustCounter(t *testing.T, id, name string, delta int64) (counter *metric.Counter) {
	t.Helper()

	counter, err := metric.NewCounter(id, name, delta)

	if err != nil {
		t.Fatalf("restore counter: %v", err)
	}

	return counter
}

type snapshotStoreStub struct {
	toLoad []repository.MetricSnapshot
	saved  []repository.MetricSnapshot

	saveErr error
	loadErr error
}

func (s *snapshotStoreStub) Save(metrics []repository.MetricSnapshot) error {
	if s.saveErr != nil {
		return s.saveErr
	}

	s.saved = append([]repository.MetricSnapshot(nil), metrics...)
	return nil
}

func (s *snapshotStoreStub) Load() ([]repository.MetricSnapshot, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}

	return append([]repository.MetricSnapshot(nil), s.toLoad...), nil
}

func TestMetricSnapshotMapper_SnapshotToDomain(t *testing.T) {
	tests := []struct {
		name        string
		snapshot    repository.MetricSnapshot
		gauge       *metric.Gauge
		counter     *metric.Counter
		errContains string
	}{
		{
			name: "gauge_success",
			snapshot: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(123.45),
			},
			gauge: mustGauge(t, "LastGC", "LastGC", 123.45),
		},
		{
			name: "counter_success",
			snapshot: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(7),
			},
			counter: mustCounter(t, "PollCount", "PollCount", 7),
		},
		{
			name: "unknown_type",
			snapshot: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  "unknown",
				Value: float64Ptr(123.45),
			},
			errContains: "unknown metric type",
		},
		{
			name: "empty_id",
			snapshot: repository.MetricSnapshot{
				ID:    "",
				Type:  TypeGauge,
				Value: float64Ptr(123.45),
			},
			errContains: "id is empty",
		},
		{
			name: "empty_value_for_gauge",
			snapshot: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: nil,
			},
			errContains: "invalid gauge snapshot",
		},
		{
			name: "empty_delta_for_counter",
			snapshot: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: nil,
			},
			errContains: "invalid counter snapshot",
		},
		{
			name: "counter_has_value_too",
			snapshot: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(1),
				Value: float64Ptr(10),
			},
			errContains: "invalid counter snapshot",
		},
	}

	// Arrange
	mapper := NewMetricSnapshotMapper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			gotGauge, gotCounter, err := mapper.SnapshotToDomain(test.snapshot)

			// Assert
			if test.errContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", test.errContains)
				}

				if !strings.Contains(err.Error(), test.errContains) {
					t.Fatalf("expected error containing %q, got %q", test.errContains, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.gauge == nil {
				if gotGauge != nil {
					t.Fatalf("expected nil gauge, got %+v", gotGauge)
				}
			} else {
				if gotGauge == nil {
					t.Fatalf("expected gauge %+v, got nil", test.gauge)
				}

				if gotGauge.Id().String() != test.gauge.Id().String() ||
					gotGauge.Name().String() != test.gauge.Name().String() ||
					gotGauge.Value() != test.gauge.Value() {
					t.Fatalf("expected gauge %+v, got %+v", test.gauge, gotGauge)
				}
			}

			if test.counter == nil {
				if gotCounter != nil {
					t.Fatalf("expected nil counter, got %+v", gotCounter)
				}
			} else {
				if gotCounter == nil {
					t.Fatalf("expected counter %+v, got nil", test.counter)
				}

				if gotCounter.Id().String() != test.counter.Id().String() ||
					gotCounter.Name().String() != test.counter.Name().String() ||
					gotCounter.Delta() != test.counter.Delta() {
					t.Fatalf("expected counter %+v, got %+v", test.counter, gotCounter)
				}
			}
		})
	}
}

func TestMetricSnapshotMapper_CounterToSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		counter     *metric.Counter
		want        repository.MetricSnapshot
		errContains string
	}{
		{
			name:    "counter_to_snapshot_success_positive_delta",
			counter: mustCounter(t, "PollCount", "PollCount", 7),
			want: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(7),
				Value: nil,
			},
		},
		{
			name:    "counter_to_snapshot_success_zero_delta",
			counter: mustCounter(t, "PollCount", "PollCount", 0),
			want: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(0),
				Value: nil,
			},
		},
		{
			name:    "counter_to_snapshot_success_negative_delta",
			counter: mustCounter(t, "PollCount", "PollCount", -5),
			want: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(-5),
				Value: nil,
			},
		},
		{
			name:        "counter_to_snapshot_nil_counter_error",
			counter:     nil,
			errContains: "counter is nil",
		},
		{
			name:    "counter_to_snapshot_uses_name_as_id",
			counter: mustCounter(t, "id-1", "PollCount", 1),
			want: repository.MetricSnapshot{
				ID:    "PollCount", // если сменишь контракт на ID, поменяй на "id-1"
				Type:  TypeCounter,
				Delta: int64Ptr(1),
				Value: nil,
			},
		},
		{
			name:    "counter_to_snapshot_max_int64",
			counter: mustCounter(t, "PollCount", "PollCount", math.MaxInt64),
			want: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(math.MaxInt64),
				Value: nil,
			},
		},
		{
			name:    "counter_to_snapshot_min_int64",
			counter: mustCounter(t, "PollCount", "PollCount", math.MinInt64),
			want: repository.MetricSnapshot{
				ID:    "PollCount",
				Type:  TypeCounter,
				Delta: int64Ptr(math.MinInt64),
				Value: nil,
			},
		},
	}

	// Arrange
	mapper := NewMetricSnapshotMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			gotSnapshot, err := mapper.CounterToSnapshot(tt.counter)

			// Assert
			if tt.errContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.errContains)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSnapshot.ID != tt.want.ID {
				t.Fatalf("expected snapshot ID %q, got %q", tt.want.ID, gotSnapshot.ID)
			}

			if gotSnapshot.Type != tt.want.Type {
				t.Fatalf("expected snapshot type %q, got %q", tt.want.Type, gotSnapshot.Type)
			}

			if (gotSnapshot.Delta == nil) != (tt.want.Delta == nil) {
				t.Fatalf("delta nil mismatch: got nil=%v want nil=%v", gotSnapshot.Delta == nil, tt.want.Delta == nil)
			}

			if gotSnapshot.Delta != nil && tt.want.Delta != nil {
				if *gotSnapshot.Delta != *tt.want.Delta {
					t.Fatalf("expected delta %d, got %d", *tt.want.Delta, *gotSnapshot.Delta)
				}
			}
		})
	}
}

func TestMetricSnapshotMapper_GaugeToSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		gauge       *metric.Gauge
		want        repository.MetricSnapshot
		errContains string
	}{
		{
			name:  "gauge_to_snapshot_success_positive_value",
			gauge: mustGauge(t, "LastGC", "LastGC", 123.45),
			want: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(123.45),
				Delta: nil,
			},
		},
		{
			name:  "gauge_to_snapshot_success_zero_value",
			gauge: mustGauge(t, "LastGC", "LastGC", 0),
			want: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(0),
				Delta: nil,
			},
		},
		{
			name:  "gauge_to_snapshot_success_negative_value",
			gauge: mustGauge(t, "LastGC", "LastGC", -1.25),
			want: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(-1.25),
				Delta: nil,
			},
		},
		{
			name:        "gauge_to_snapshot_nil_gauge_error",
			gauge:       nil,
			errContains: "gauge is nil",
		},
		{
			name:  "gauge_to_snapshot_uses_name_as_id",
			gauge: mustGauge(t, "id-1", "LastGC", 1.0),
			want: repository.MetricSnapshot{
				ID:    "LastGC", // если сменишь контракт на ID, поменяй на "id-1"
				Type:  TypeGauge,
				Value: float64Ptr(1.0),
				Delta: nil,
			},
		},
		{
			name:  "gauge_to_snapshot_max_float64",
			gauge: mustGauge(t, "LastGC", "LastGC", math.MaxFloat64),
			want: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(math.MaxFloat64),
				Delta: nil,
			},
		},
		{
			name:  "gauge_to_snapshot_smallest_nonzero_float64",
			gauge: mustGauge(t, "LastGC", "LastGC", math.SmallestNonzeroFloat64),
			want: repository.MetricSnapshot{
				ID:    "LastGC",
				Type:  TypeGauge,
				Value: float64Ptr(math.SmallestNonzeroFloat64),
				Delta: nil,
			},
		},
	}

	// Arrange
	mapper := NewMetricSnapshotMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			gotSnapshot, err := mapper.GaugeToSnapshot(tt.gauge)

			// Assert
			if tt.errContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.errContains)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSnapshot.ID != tt.want.ID {
				t.Fatalf("expected snapshot ID %q, got %q", tt.want.ID, gotSnapshot.ID)
			}

			if gotSnapshot.Type != tt.want.Type {
				t.Fatalf("expected snapshot type %q, got %q", tt.want.Type, gotSnapshot.Type)
			}

			if (gotSnapshot.Value == nil) != (tt.want.Value == nil) {
				t.Fatalf("delta nil mismatch: got nil=%v want nil=%v", gotSnapshot.Value == nil, tt.want.Value == nil)
			}

			if gotSnapshot.Value != nil && tt.want.Value != nil {
				if *gotSnapshot.Value != *tt.want.Value {
					t.Fatalf("expected value %v, got %v", *tt.want.Value, *gotSnapshot.Value)
				}
			}
		})
	}
}
