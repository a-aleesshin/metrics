package memory

import (
	"testing"

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

				if err := storage.SaveGauge(gauge); err != nil {
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
			got, err := storage.GetGaugeByName(name)

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

				if err := storage.SaveCounter(counter); err != nil {
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
			got, err := storage.GetCounterByName(name)

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
				_ = s.SaveGauge(g)
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
			gotValue, gotFound, err := s.FindGaugeByName(name)

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
				_ = s.SaveCounter(c)
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
			gotDelta, gotFound, err := s.FindCounterByName(name)

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
				_ = s.SaveGauge(g1)
				_ = s.SaveGauge(g2)
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
			got, err := s.ListGauges()

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
				_ = s.SaveCounter(c1)
				_ = s.SaveCounter(c2)
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
			got, err := s.ListCounters()

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
