package mapper

import (
	"math"
	"testing"

	appdto "github.com/a-aleesshin/metrics/internal/agent/application/dto"
)

func TestToSendMetric(t *testing.T) {
	tests := []struct {
		name      string
		input     appdto.MetricDTO
		wantID    string
		wantType  string
		wantValue *float64
		wantDelta *int64
		wantErr   bool
	}{
		{
			name: "maps_gauge",
			input: appdto.MetricDTO{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "123.45",
			},
			wantID:    "Alloc",
			wantType:  "gauge",
			wantValue: float64Ptr(123.45),
		},
		{
			name: "maps_counter",
			input: appdto.MetricDTO{
				Type:  "counter",
				Name:  "PollCount",
				Value: "7",
			},
			wantID:    "PollCount",
			wantType:  "counter",
			wantDelta: int64Ptr(7),
		},
		{
			name: "normalizes_nan_gauge_to_zero",
			input: appdto.MetricDTO{
				Type:  "gauge",
				Name:  "RandomValue",
				Value: "NaN",
			},
			wantID:    "RandomValue",
			wantType:  "gauge",
			wantValue: float64Ptr(0),
		},
		{
			name: "normalizes_positive_infinity_gauge_to_zero",
			input: appdto.MetricDTO{
				Type:  "gauge",
				Name:  "RandomValue",
				Value: "+Inf",
			},
			wantID:    "RandomValue",
			wantType:  "gauge",
			wantValue: float64Ptr(0),
		},
		{
			name: "invalid_gauge_value_returns_error",
			input: appdto.MetricDTO{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "not-float",
			},
			wantErr: true,
		},
		{
			name: "invalid_counter_value_returns_error",
			input: appdto.MetricDTO{
				Type:  "counter",
				Name:  "PollCount",
				Value: "1.5",
			},
			wantErr: true,
		},
		{
			name: "unsupported_type_returns_error",
			input: appdto.MetricDTO{
				Type:  "histogram",
				Name:  "Buckets",
				Value: "1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := ToSendMetric(tt.input)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.ID != tt.wantID {
				t.Fatalf("expected ID %q, got %q", tt.wantID, got.ID)
			}

			if got.MType != tt.wantType {
				t.Fatalf("expected type %q, got %q", tt.wantType, got.MType)
			}

			assertFloat64PtrEqual(t, tt.wantValue, got.Value)
			assertInt64PtrEqual(t, tt.wantDelta, got.Delta)
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func assertFloat64PtrEqual(t *testing.T, want, got *float64) {
	t.Helper()

	if want == nil && got == nil {
		return
	}

	if want == nil || got == nil {
		t.Fatalf("expected value %v, got %v", want, got)
	}

	if math.Abs(*want-*got) > 0.000000001 {
		t.Fatalf("expected value %v, got %v", *want, *got)
	}
}

func assertInt64PtrEqual(t *testing.T, want, got *int64) {
	t.Helper()

	if want == nil && got == nil {
		return
	}

	if want == nil || got == nil {
		t.Fatalf("expected delta %v, got %v", want, got)
	}

	if *want != *got {
		t.Fatalf("expected delta %d, got %d", *want, *got)
	}
}
