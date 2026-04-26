package file

import (
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

func f64(v float64) *float64 {
	return &v
}

func i64(v int64) *int64 {
	return &v
}

func TestSnapshotStore_SaveLoad(t *testing.T) {
	tests := []struct {
		name    string
		input   []repository.MetricSnapshot
		wantLen int
		wantErr bool
	}{
		{
			name: "round trip",
			input: []repository.MetricSnapshot{
				{ID: "Alloc", Type: "gauge", Value: f64(12.3)},
				{ID: "PollCount", Type: "counter", Delta: i64(7)},
			},
			wantLen: 2,
		},
		{
			name:    "empty slice saved and loaded",
			input:   []repository.MetricSnapshot{},
			wantLen: 0,
		},
		{
			name: "invalid snapshot type",
			input: []repository.MetricSnapshot{
				{ID: "X", Type: "unknown"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			path := t.TempDir() + "/metrics.json"
			store, err := NewSnapshotStore(path)
			if err != nil {
				t.Fatalf("new store: %v", err)
			}

			// Act
			err = store.Save(tt.input)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected save error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("save failed: %v", err)
			}

			got, err := store.Load()

			if err != nil {
				t.Fatalf("load failed: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf("expected len=%d, got=%d", tt.wantLen, len(got))
			}
		})
	}
}

func TestSnapshotStore_Load_FileNotExists_ReturnsEmpty(t *testing.T) {
	// Arrange
	path := t.TempDir() + "/not-exists.json"
	store, err := NewSnapshotStore(path)

	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	// Act
	got, err := store.Load()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d", len(got))
	}
}
