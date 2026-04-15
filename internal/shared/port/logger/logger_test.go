package logger

import (
	"errors"
	"testing"
)

func TestFieldHelpers(t *testing.T) {
	tests := []struct {
		name     string
		build    func() Field
		wantKey  string
		wantType string
		wantAny  any
	}{
		{
			name: "String helper",
			build: func() Field {
				return String("metric", "Alloc")
			},
			wantKey:  "metric",
			wantType: "string",
			wantAny:  "Alloc",
		},
		{
			name: "Int helper",
			build: func() Field {
				return Int("delta", 7)
			},
			wantKey:  "delta",
			wantType: "int",
			wantAny:  7,
		},
		{
			name: "Err helper with error",
			build: func() Field {
				return Err(errors.New("boom"))
			},
			wantKey:  "error",
			wantType: "string",
			wantAny:  "boom",
		},
		{
			name: "Err helper with nil",
			build: func() Field {
				return Err(nil)
			},
			wantKey:  "error",
			wantType: "nil",
			wantAny:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// (нет отдельной подготовки)

			// Act
			got := tt.build()

			// Assert
			if got.Key != tt.wantKey {
				t.Fatalf("expected key %q, got %q", tt.wantKey, got.Key)
			}

			switch tt.wantType {
			case "string":
				v, ok := got.Value.(string)
				if !ok {
					t.Fatalf("expected string value, got %T", got.Value)
				}
				if v != tt.wantAny.(string) {
					t.Fatalf("expected value %q, got %q", tt.wantAny.(string), v)
				}
			case "int":
				v, ok := got.Value.(int)
				if !ok {
					t.Fatalf("expected int value, got %T", got.Value)
				}
				if v != tt.wantAny.(int) {
					t.Fatalf("expected value %d, got %d", tt.wantAny.(int), v)
				}
			case "nil":
				if got.Value != nil {
					t.Fatalf("expected nil value, got %v", got.Value)
				}
			default:
				t.Fatalf("unknown wantType: %s", tt.wantType)
			}
		})
	}
}
