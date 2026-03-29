package metric

import (
	"errors"
	"testing"
)

func TestNewGauge(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		metricName string
		value      float64
		wantErr    error
		wantID     ID
		wantName   Name
		wantValue  float64
	}{
		{
			name:       "valid gauge",
			id:         "gauge-id",
			metricName: "Alloc",
			value:      123.45,
			wantErr:    nil,
			wantID:     ID("gauge-id"),
			wantName:   Name("Alloc"),
			wantValue:  123.45,
		},
		{
			name:       "empty id",
			id:         "",
			metricName: "Alloc",
			value:      123.45,
			wantErr:    ErrIDEmpty,
		},
		{
			name:       "empty name",
			id:         "gauge-id",
			metricName: "",
			value:      123.45,
			wantErr:    ErrNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange

			// Act
			got, err := NewGauge(tt.id, tt.metricName, tt.value)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Id() != tt.wantID {
				t.Fatalf("expected id %q, got %q", tt.wantID, got.Id())
			}

			if got.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name())
			}

			if got.Value() != tt.wantValue {
				t.Fatalf("expected value %v, got %v", tt.wantValue, got.Value())
			}
		})
	}
}

func TestGauge_Rename(t *testing.T) {
	tests := []struct {
		name        string
		initialID   string
		initialName string
		newName     string
		wantErr     error
		wantName    Name
	}{
		{
			name:        "rename success",
			initialID:   "gauge-id",
			initialName: "Alloc",
			newName:     "HeapAlloc",
			wantErr:     nil,
			wantName:    Name("HeapAlloc"),
		},
		{
			name:        "rename empty name",
			initialID:   "gauge-id",
			initialName: "Alloc",
			newName:     "",
			wantErr:     ErrNameEmpty,
			wantName:    Name("Alloc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			gauge, err := NewGauge(tt.initialID, tt.initialName, 10)
			if err != nil {
				t.Fatalf("unexpected setup error: %v", err)
			}

			// Act
			err = gauge.Rename(tt.newName)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if gauge.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, gauge.Name())
			}
		})
	}
}

func TestGauge_UpdateValue(t *testing.T) {
	// Arrange
	gauge, err := NewGauge("gauge-id", "Alloc", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Act
	gauge.UpdateValue(20.5)

	// Assert
	if gauge.Value() != 20.5 {
		t.Fatalf("expected value %v, got %v", 20.5, gauge.Value())
	}
}

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		metricName string
		delta      int64
		wantErr    error
		wantID     ID
		wantName   Name
		wantDelta  int64
	}{
		{
			name:       "valid counter",
			id:         "counter-id",
			metricName: "PollCount",
			delta:      1,
			wantErr:    nil,
			wantID:     ID("counter-id"),
			wantName:   Name("PollCount"),
			wantDelta:  1,
		},
		{
			name:       "empty id",
			id:         "",
			metricName: "PollCount",
			delta:      1,
			wantErr:    ErrIDEmpty,
		},
		{
			name:       "empty name",
			id:         "counter-id",
			metricName: "",
			delta:      1,
			wantErr:    ErrNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange

			// Act
			got, err := NewCounter(tt.id, tt.metricName, tt.delta)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Id() != tt.wantID {
				t.Fatalf("expected id %q, got %q", tt.wantID, got.Id())
			}

			if got.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name())
			}

			if got.Delta() != tt.wantDelta {
				t.Fatalf("expected delta %d, got %d", tt.wantDelta, got.Delta())
			}
		})
	}
}

func TestCounter_Rename(t *testing.T) {
	tests := []struct {
		name        string
		initialID   string
		initialName string
		newName     string
		wantErr     error
		wantName    Name
	}{
		{
			name:        "rename success",
			initialID:   "counter-id",
			initialName: "PollCount",
			newName:     "TotalCount",
			wantErr:     nil,
			wantName:    Name("TotalCount"),
		},
		{
			name:        "rename empty name",
			initialID:   "counter-id",
			initialName: "PollCount",
			newName:     "",
			wantErr:     ErrNameEmpty,
			wantName:    Name("PollCount"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			counter, err := NewCounter(tt.initialID, tt.initialName, 1)
			if err != nil {
				t.Fatalf("unexpected setup error: %v", err)
			}

			// Act
			err = counter.Rename(tt.newName)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if counter.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, counter.Name())
			}
		})
	}
}

func TestCounter_Add(t *testing.T) {
	// Arrange
	counter, err := NewCounter("counter-id", "PollCount", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Act
	counter.Add(5)

	// Assert
	if counter.Delta() != 15 {
		t.Fatalf("expected delta %d, got %d", 15, counter.Delta())
	}
}
