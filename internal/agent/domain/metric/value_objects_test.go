package metric

import (
	"errors"
	"testing"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Name
		wantErr error
	}{
		{
			name:    "valid name",
			input:   "Alloc",
			want:    Name("Alloc"),
			wantErr: nil,
		},
		{
			name:    "empty name",
			input:   "",
			want:    "",
			wantErr: ErrNameEmpty,
		},
		{
			name:    "name with underscore",
			input:   "test_counter",
			want:    Name("test_counter"),
			wantErr: nil,
		},
		{
			name:    "name with mixed case",
			input:   "HeapAlloc",
			want:    Name("HeapAlloc"),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := NewName(tt.input)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestName_String(t *testing.T) {
	// Arrange
	name := Name("Alloc")

	// Act
	got := name.String()

	// Assert
	if got != "Alloc" {
		t.Fatalf("expected %q, got %q", "Alloc", got)
	}
}
