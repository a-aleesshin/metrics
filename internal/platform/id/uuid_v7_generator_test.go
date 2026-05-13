package id

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDV7Generator_NewID(t *testing.T) {
	// Arrange
	generator := NewUUIDV7Generator()

	// Act
	got, err := generator.NewID()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.String() == "" {
		t.Fatal("expected id not to be empty")
	}

	parsed, err := uuid.Parse(got.String())
	if err != nil {
		t.Fatalf("expected valid uuid, got %q: %v", got.String(), err)
	}

	if parsed.Version() != 7 {
		t.Fatalf("expected uuid version 7, got %d", parsed.Version())
	}
}

func TestUUIDV7Generator_NewID_ReturnsDifferentIDs(t *testing.T) {
	// Arrange
	generator := NewUUIDV7Generator()

	// Act
	first, err := generator.NewID()
	if err != nil {
		t.Fatalf("unexpected first error: %v", err)
	}

	second, err := generator.NewID()
	if err != nil {
		t.Fatalf("unexpected second error: %v", err)
	}

	// Assert
	if first == second {
		t.Fatalf("expected different ids, got %q", first.String())
	}
}
