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
