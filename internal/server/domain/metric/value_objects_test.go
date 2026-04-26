package metric

import (
	"errors"
	"testing"
)

func TestNewID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ID
		wantErr error
	}{
		{
			name:    "valid_id",
			input:   "metric-id",
			want:    ID("metric-id"),
			wantErr: nil,
		},
		{
			name:    "empty_id",
			input:   "",
			want:    "",
			wantErr: ErrIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewID(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Name
		wantErr error
	}{
		{
			name:    "valid_name",
			input:   "Alloc",
			want:    Name("Alloc"),
			wantErr: nil,
		},
		{
			name:    "empty_name",
			input:   "",
			want:    "",
			wantErr: ErrNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewName(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
