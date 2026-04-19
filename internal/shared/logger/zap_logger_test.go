package logger

import (
	"testing"

	portlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"go.uber.org/zap/zapcore"
)

func TestZapFields(t *testing.T) {
	tests := []struct {
		name   string
		input  []portlogger.Field
		wantKV map[string]any
	}{
		{
			name:   "empty fields",
			input:  []portlogger.Field{},
			wantKV: map[string]any{},
		},
		{
			name: "maps all fields",
			input: []portlogger.Field{
				{Key: "name", Value: "Alloc"},
				{Key: "delta", Value: 7},
				{Key: "ok", Value: true},
			},
			wantKV: map[string]any{
				"name":  "Alloc",
				"delta": int64(7),
				"ok":    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := zapFields(tt.input)

			// Assert count
			if len(got) != len(tt.input) {
				t.Fatalf("expected %d fields, got %d", len(tt.input), len(got))
			}

			enc := zapcore.NewMapObjectEncoder()
			for _, f := range got {
				f.AddTo(enc)
			}

			if len(enc.Fields) != len(tt.wantKV) {
				t.Fatalf("expected %d encoded fields, got %d", len(tt.wantKV), len(enc.Fields))
			}

			for k, want := range tt.wantKV {
				gotV, ok := enc.Fields[k]
				if !ok {
					t.Fatalf("expected key %q not found", k)
				}
				if gotV != want {
					t.Fatalf("key %q: expected %v, got %v", k, want, gotV)
				}
			}
		})
	}
}
