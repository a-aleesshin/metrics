package memory

import (
	"fmt"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

func TestMemStorage_GaugeOperations(t *testing.T) {
	tests := []struct {
		name       string
		prepare    func(t *testing.T, storage *MemStorage)
		lookupName string
		wantNil    bool
		wantID     string
		wantName   string
		wantValue  float64
	}{
		{
			name: "save and get gauge",
			prepare: func(t *testing.T, storage *MemStorage) {
				gauge, err := metric.NewGauge("gauge-id", "Alloc", 123.45)
				if err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}

				if err := storage.SaveGauge(t.Context(), gauge); err != nil {
					t.Fatalf("unexpected save error: %v", err)
				}
			},
			lookupName: "Alloc",
			wantNil:    false,
			wantID:     "gauge-id",
			wantName:   "Alloc",
			wantValue:  123.45,
		},
		{
			name:       "get missing gauge returns nil",
			prepare:    func(t *testing.T, storage *MemStorage) {},
			lookupName: "MissingGauge",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			storage := NewMemStorage()
			tt.prepare(t, storage)

			name, err := metric.NewName(tt.lookupName)
			if err != nil && !tt.wantNil {
				t.Fatalf("unexpected name error: %v", err)
			}

			// Act
			got, err := storage.GetGaugeByName(t.Context(), name)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil gauge, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected gauge, got nil")
			}

			if got.Id().String() != tt.wantID {
				t.Fatalf("expected id %q, got %q", tt.wantID, got.Id().String())
			}

			if got.Name().String() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name().String())
			}

			if got.Value() != tt.wantValue {
				t.Fatalf("expected value %v, got %v", tt.wantValue, got.Value())
			}
		})
	}
}

func TestMemStorage_CounterOperations(t *testing.T) {
	tests := []struct {
		name       string
		prepare    func(t *testing.T, storage *MemStorage)
		lookupName string
		wantNil    bool
		wantID     string
		wantName   string
		wantDelta  int64
	}{
		{
			name: "save and get counter",
			prepare: func(t *testing.T, storage *MemStorage) {
				counter, err := metric.NewCounter("counter-id", "PollCount", 7)
				if err != nil {
					t.Fatalf("unexpected setup error: %v", err)
				}

				if err := storage.SaveCounter(t.Context(), counter); err != nil {
					t.Fatalf("unexpected save error: %v", err)
				}
			},
			lookupName: "PollCount",
			wantNil:    false,
			wantID:     "counter-id",
			wantName:   "PollCount",
			wantDelta:  7,
		},
		{
			name:       "get missing counter returns nil",
			prepare:    func(t *testing.T, storage *MemStorage) {},
			lookupName: "MissingCounter",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			storage := NewMemStorage()
			tt.prepare(t, storage)

			name, err := metric.NewName(tt.lookupName)
			if err != nil && !tt.wantNil {
				t.Fatalf("unexpected name error: %v", err)
			}

			// Act
			got, err := storage.GetCounterByName(t.Context(), name)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil counter, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected counter, got nil")
			}

			if got.Id().String() != tt.wantID {
				t.Fatalf("expected id %q, got %q", tt.wantID, got.Id().String())
			}

			if got.Name().String() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name().String())
			}

			if got.Delta() != tt.wantDelta {
				t.Fatalf("expected delta %d, got %d", tt.wantDelta, got.Delta())
			}
		})
	}
}

func TestMemStorage_FindGaugeByName(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, s *MemStorage)
		lookup    string
		wantFound bool
		wantValue float64
	}{
		{
			name:      "not found",
			prepare:   func(t *testing.T, s *MemStorage) {},
			lookup:    "Missing",
			wantFound: false,
		},
		{
			name: "found",
			prepare: func(t *testing.T, s *MemStorage) {
				g, _ := metric.NewGauge("g1", "Alloc", 123.45)
				_ = s.SaveGauge(t.Context(), g)
			},
			lookup:    "Alloc",
			wantFound: true,
			wantValue: 123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)
			name, _ := metric.NewName(tt.lookup)

			// Act
			gotValue, gotFound, err := s.FindGaugeByName(t.Context(), name)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFound != tt.wantFound {
				t.Fatalf("expected found=%v, got %v", tt.wantFound, gotFound)
			}
			if gotFound && gotValue != tt.wantValue {
				t.Fatalf("expected value %v, got %v", tt.wantValue, gotValue)
			}
		})
	}
}

func TestMemStorage_FindCounterByName(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, s *MemStorage)
		lookup    string
		wantFound bool
		wantDelta int64
	}{
		{
			name:      "not found",
			prepare:   func(t *testing.T, s *MemStorage) {},
			lookup:    "Missing",
			wantFound: false,
		},
		{
			name: "found",
			prepare: func(t *testing.T, s *MemStorage) {
				c, _ := metric.NewCounter("c1", "PollCount", 7)
				_ = s.SaveCounter(t.Context(), c)
			},
			lookup:    "PollCount",
			wantFound: true,
			wantDelta: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)
			name, _ := metric.NewName(tt.lookup)

			// Act
			gotDelta, gotFound, err := s.FindCounterByName(t.Context(), name)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFound != tt.wantFound {
				t.Fatalf("expected found=%v, got %v", tt.wantFound, gotFound)
			}
			if gotFound && gotDelta != tt.wantDelta {
				t.Fatalf("expected delta %d, got %d", tt.wantDelta, gotDelta)
			}
		})
	}
}

func TestMemStorage_ListGauges(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, s *MemStorage)
		wantCount int
		wantByKey map[string]float64
	}{
		{
			name:      "empty storage",
			prepare:   func(t *testing.T, s *MemStorage) {},
			wantCount: 0,
			wantByKey: map[string]float64{},
		},
		{
			name: "returns all gauges",
			prepare: func(t *testing.T, s *MemStorage) {
				g1, _ := metric.NewGauge("g1", "Alloc", 10.5)
				g2, _ := metric.NewGauge("g2", "HeapInuse", 42)
				_ = s.SaveGauge(t.Context(), g1)
				_ = s.SaveGauge(t.Context(), g2)
			},
			wantCount: 2,
			wantByKey: map[string]float64{
				"Alloc":     10.5,
				"HeapInuse": 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)

			// Act
			got, err := s.ListGauges(t.Context())

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Fatalf("expected %d gauges, got %d", tt.wantCount, len(got))
			}

			gotByKey := make(map[string]float64, len(got))
			for _, g := range got {
				gotByKey[g.Name] = g.Value
			}

			if len(gotByKey) != len(tt.wantByKey) {
				t.Fatalf("expected %d unique gauges, got %d", len(tt.wantByKey), len(gotByKey))
			}
			for name, wantValue := range tt.wantByKey {
				gotValue, ok := gotByKey[name]
				if !ok {
					t.Fatalf("expected gauge %q not found", name)
				}
				if gotValue != wantValue {
					t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}
		})
	}
}

func TestMemStorage_ListCounters(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T, s *MemStorage)
		wantCount int
		wantByKey map[string]int64
	}{
		{
			name:      "empty storage",
			prepare:   func(t *testing.T, s *MemStorage) {},
			wantCount: 0,
			wantByKey: map[string]int64{},
		},
		{
			name: "returns all counters",
			prepare: func(t *testing.T, s *MemStorage) {
				c1, _ := metric.NewCounter("c1", "PollCount", 3)
				c2, _ := metric.NewCounter("c2", "Requests", 9)
				_ = s.SaveCounter(t.Context(), c1)
				_ = s.SaveCounter(t.Context(), c2)
			},
			wantCount: 2,
			wantByKey: map[string]int64{
				"PollCount": 3,
				"Requests":  9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)

			// Act
			got, err := s.ListCounters(t.Context())

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Fatalf("expected %d counters, got %d", tt.wantCount, len(got))
			}

			gotByKey := make(map[string]int64, len(got))
			for _, c := range got {
				gotByKey[c.Name] = c.Delta
			}

			if len(gotByKey) != len(tt.wantByKey) {
				t.Fatalf("expected %d unique counters, got %d", len(tt.wantByKey), len(gotByKey))
			}
			for name, wantDelta := range tt.wantByKey {
				gotDelta, ok := gotByKey[name]
				if !ok {
					t.Fatalf("expected counter %q not found", name)
				}
				if gotDelta != wantDelta {
					t.Fatalf("counter %q: expected %d, got %d", name, wantDelta, gotDelta)
				}
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name         string
		prepare      func(t *testing.T, s *MemStorage)
		wantGaugeLen int
		wantCtrLen   int
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			name: "empty_storage",
			prepare: func(t *testing.T, s *MemStorage) {
			},
			wantGauges:   map[string]float64{},
			wantCounters: map[string]int64{},
		},
		{
			name: "returns_all_metrics",
			prepare: func(t *testing.T, s *MemStorage) {
				g1, _ := metric.NewGauge("g1", "Alloc", 10.5)
				g2, _ := metric.NewGauge("g2", "HeapInuse", 42)
				c1, _ := metric.NewCounter("c1", "PollCount", 3)
				c2, _ := metric.NewCounter("c2", "Requests", 9)
				_ = s.SaveGauge(t.Context(), g1)
				_ = s.SaveGauge(t.Context(), g2)
				_ = s.SaveCounter(t.Context(), c1)
				_ = s.SaveCounter(t.Context(), c2)
			},
			wantGaugeLen: 2,
			wantCtrLen:   2,
			wantGauges: map[string]float64{
				"Alloc":     10.5,
				"HeapInuse": 42,
			},
			wantCounters: map[string]int64{
				"PollCount": 3,
				"Requests":  9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)

			// Act
			got, err := s.GetAllMetrics(t.Context())

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got.Gauges) != tt.wantGaugeLen {
				t.Fatalf("expected %d gauges, got %d", tt.wantGaugeLen, len(got.Gauges))
			}

			if len(got.Counters) != tt.wantCtrLen {
				t.Fatalf("expected %d counters, got %d", tt.wantCtrLen, len(got.Counters))
			}

			gotGauges := make(map[string]float64, len(got.Gauges))

			for _, g := range got.Gauges {
				if g == nil {
					fmt.Println("nil gauge")
				}

				gotGauges[g.Name().String()] = g.Value()
			}

			gotCounters := make(map[string]int64, len(got.Counters))
			for _, c := range got.Counters {
				if c == nil {
					t.Fatal("got nil counter")
				}
				gotCounters[c.Name().String()] = c.Delta()
			}

			if len(gotGauges) != len(tt.wantGauges) {
				t.Fatalf("expected %d unique gauges, got %d", len(tt.wantGauges), len(gotGauges))
			}

			for name, wantValue := range tt.wantGauges {
				gotValue, ok := gotGauges[name]

				if !ok {
					t.Fatalf("expected gauge %q not found", name)
				}

				if gotValue != wantValue {
					t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}

			for name, wantDelta := range tt.wantCounters {
				gotDelta, ok := gotCounters[name]

				if !ok {
					t.Fatalf("expected counter %q not found", name)
				}

				if gotDelta != wantDelta {
					t.Fatalf("counter %q: expected %d, got %d", name, wantDelta, gotDelta)
				}
			}
		})
	}
}

func TestMemStorage_UpdateBatch(t *testing.T) {
	tests := []struct {
		name         string
		prepare      func(t *testing.T, s *MemStorage)
		batch        func(t *testing.T) repository.MetricBatch
		wantGauges   map[string]float64
		wantCounters map[string]int64
	}{
		{
			name:    "empty batch does nothing",
			prepare: func(t *testing.T, s *MemStorage) {},
			batch: func(t *testing.T) repository.MetricBatch {
				return repository.MetricBatch{}
			},
			wantGauges:   map[string]float64{},
			wantCounters: map[string]int64{},
		},
		{
			name:    "adds new gauges and counters",
			prepare: func(t *testing.T, s *MemStorage) {},
			batch: func(t *testing.T) repository.MetricBatch {
				g, err := metric.NewGauge("g1", "Alloc", 123.45)
				if err != nil {
					t.Fatalf("unexpected gauge error: %v", err)
				}

				c, err := metric.NewCounter("c1", "PollCount", 7)
				if err != nil {
					t.Fatalf("unexpected counter error: %v", err)
				}

				return repository.MetricBatch{
					Gauges:   []*metric.Gauge{g},
					Counters: []*metric.Counter{c},
				}
			},
			wantGauges: map[string]float64{
				"Alloc": 123.45,
			},
			wantCounters: map[string]int64{
				"PollCount": 7,
			},
		},
		{
			name: "replaces gauge and increments counter",
			prepare: func(t *testing.T, s *MemStorage) {
				g, _ := metric.NewGauge("old-gauge-id", "Alloc", 10)
				c, _ := metric.NewCounter("old-counter-id", "PollCount", 3)

				if err := s.SaveGauge(t.Context(), g); err != nil {
					t.Fatalf("unexpected save gauge error: %v", err)
				}

				if err := s.SaveCounter(t.Context(), c); err != nil {
					t.Fatalf("unexpected save counter error: %v", err)
				}
			},
			batch: func(t *testing.T) repository.MetricBatch {
				g, err := metric.NewGauge("new-gauge-id", "Alloc", 20)
				if err != nil {
					t.Fatalf("unexpected gauge error: %v", err)
				}

				c, err := metric.NewCounter("new-counter-id", "PollCount", 5)
				if err != nil {
					t.Fatalf("unexpected counter error: %v", err)
				}

				return repository.MetricBatch{
					Gauges:   []*metric.Gauge{g},
					Counters: []*metric.Counter{c},
				}
			},
			wantGauges: map[string]float64{
				"Alloc": 20,
			},
			wantCounters: map[string]int64{
				"PollCount": 8,
			},
		},
		{
			name:    "increments duplicate counters in same batch",
			prepare: func(t *testing.T, s *MemStorage) {},
			batch: func(t *testing.T) repository.MetricBatch {
				c1, _ := metric.NewCounter("c1", "PollCount", 3)
				c2, _ := metric.NewCounter("c2", "PollCount", 5)

				return repository.MetricBatch{
					Counters: []*metric.Counter{c1, c2},
				}
			},
			wantGauges: map[string]float64{},
			wantCounters: map[string]int64{
				"PollCount": 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := NewMemStorage()
			tt.prepare(t, s)

			// Act
			err := s.UpdateBatch(t.Context(), tt.batch(t))

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			state, err := s.GetAllMetrics(t.Context())
			if err != nil {
				t.Fatalf("unexpected get all metrics error: %v", err)
			}

			gotGauges := make(map[string]float64, len(state.Gauges))
			for _, g := range state.Gauges {
				gotGauges[g.Name().String()] = g.Value()
			}

			gotCounters := make(map[string]int64, len(state.Counters))
			for _, c := range state.Counters {
				gotCounters[c.Name().String()] = c.Delta()
			}

			if len(gotGauges) != len(tt.wantGauges) {
				t.Fatalf("expected %d gauges, got %d", len(tt.wantGauges), len(gotGauges))
			}

			for name, wantValue := range tt.wantGauges {
				gotValue, ok := gotGauges[name]
				if !ok {
					t.Fatalf("expected gauge %q not found", name)
				}
				if gotValue != wantValue {
					t.Fatalf("gauge %q: expected %v, got %v", name, wantValue, gotValue)
				}
			}

			if len(gotCounters) != len(tt.wantCounters) {
				t.Fatalf("expected %d counters, got %d", len(tt.wantCounters), len(gotCounters))
			}

			for name, wantDelta := range tt.wantCounters {
				gotDelta, ok := gotCounters[name]
				if !ok {
					t.Fatalf("expected counter %q not found", name)
				}
				if gotDelta != wantDelta {
					t.Fatalf("counter %q: expected %d, got %d", name, wantDelta, gotDelta)
				}
			}
		})
	}
}
