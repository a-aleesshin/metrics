package metric

import (
	"testing"
)

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName Name
		value      int64
		wantName   Name
		wantValue  int64
	}{
		{
			name:       "valid counter",
			metricName: Name("test_counter"),
			value:      10,
			wantName:   Name("test_counter"),
			wantValue:  10,
		},
		{
			name:       "negative value",
			metricName: Name("test_counter"),
			value:      -10,
			wantName:   Name("test_counter"),
			wantValue:  -10,
		},
		{
			name:       "zero value",
			metricName: Name("test_counter"),
			value:      0,
			wantName:   Name("test_counter"),
			wantValue:  0,
		},
		{
			name:       "big value",
			metricName: Name("test_counter"),
			value:      10034325543436,
			wantName:   Name("test_counter"),
			wantValue:  10034325543436,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := NewCounter(tt.metricName, tt.value)

			// Assert
			if got.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name())
			}

			if got.Value() != tt.wantValue {
				t.Fatalf("expected value %d, got %d", tt.wantValue, got.Value())
			}
		})
	}
}

func TestNewGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName Name
		value      float64
		wantName   Name
		wantValue  float64
	}{
		{
			name:       "valid gauge",
			metricName: Name("alloc"),
			value:      10.5,
			wantName:   Name("alloc"),
			wantValue:  10.5,
		},
		{
			name:       "zero value",
			metricName: Name("alloc"),
			value:      0,
			wantName:   Name("alloc"),
			wantValue:  0,
		},
		{
			name:       "negative value",
			metricName: Name("temperature"),
			value:      -12.3,
			wantName:   Name("temperature"),
			wantValue:  -12.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := NewGauge(tt.metricName, tt.value)

			// Assert
			if got.Name() != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.Name())
			}

			if got.Value() != tt.wantValue {
				t.Fatalf("expected value %v, got %v", tt.wantValue, got.Value())
			}
		})
	}
}
