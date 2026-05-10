package health

import (
	"context"
	"errors"
	"testing"
)

type checkStub struct {
	name string
	err  error
}

func (check checkStub) Name() string {
	return check.name
}

func (check checkStub) Check(ctx context.Context) error {
	return check.err
}

func TestService_AllCheck_Ok(t *testing.T) {
	// Arrange
	service := NewService(
		checkStub{
			name: "postgres",
			err:  nil,
		},
		checkStub{
			name: "mysql",
			err:  nil,
		},
		checkStub{
			name: "file-storage",
			err:  nil,
		},
		checkStub{
			name: "redis",
			err:  nil,
		},
		checkStub{
			name: "iam-service",
			err:  nil,
		},
	)

	want := []CheckResult{
		{Name: "postgres", Status: "ok"},
		{Name: "mysql", Status: "ok"},
		{Name: "file-storage", Status: "ok"},
		{Name: "redis", Status: "ok"},
		{Name: "iam-service", Status: "ok"},
	}

	// Act
	result := service.Check(context.Background())

	// Assert
	if result.Status != "ok" {
		t.Fatalf("expected status ok, got %s", result.Status)
	}

	for i, check := range result.Checks {
		if check.Status != "ok" {
			t.Fatalf("expected status ok for check %s, got %s", check.Name, check.Status)
		}

		if check.Name != want[i].Name {
			t.Fatalf("expected check %s, got %s", want[i].Name, check.Name)
		}

		if check.Status != want[i].Status {
			t.Fatalf("expected status %s for check %s, got %s", want[i].Status, check.Name, check.Status)
		}
	}
}

func TestService_Unhealthy(t *testing.T) {
	// Arrange
	service := NewService(
		checkStub{
			name: "postgres",
			err:  nil,
		},
		checkStub{
			name: "mysql",
			err:  nil,
		},
		checkStub{
			name: "file-storage",
			err:  errors.New("file storage error"),
		},
		checkStub{
			name: "redis",
			err:  nil,
		},
		checkStub{
			name: "iam-service",
			err:  nil,
		},
	)

	want := []CheckResult{
		{Name: "postgres", Status: "ok"},
		{Name: "mysql", Status: "ok"},
		{Name: "file-storage", Status: "error", Error: "file storage error"},
		{Name: "redis", Status: "ok"},
		{Name: "iam-service", Status: "ok"},
	}

	// Act
	result := service.Check(context.Background())

	// Assert
	if result.Status != "unhealthy" {
		t.Fatalf("expected status error, got %s", result.Status)
	}

	for i, check := range result.Checks {
		if check.Status == "error" && check.Error != want[i].Error {
			t.Fatalf("expected error %s, got %s", want[i].Error, check.Error)
		}

		if check.Status == "ok" && check.Status != want[i].Status {
			t.Fatalf("expected status %s, got %s", want[i].Status, check.Status)
		}

		if check.Name != want[i].Name {
			t.Fatalf("expected check %s, got %s", want[i].Name, check.Name)
		}

		if check.Status != want[i].Status {
			t.Fatalf("expected status %s for check %s, got %s", want[i].Status, check.Name, check.Status)
		}
	}
}

func TestService_EmptyChecks(t *testing.T) {
	// Arrange
	service := NewService()

	// Act
	result := service.Check(context.Background())

	// Assert
	if result.Status != "ok" {
		t.Fatalf("expected status ok, got %s", result.Status)
	}

	if len(result.Checks) != 0 {
		t.Fatalf("expected empty checks, got %d", len(result.Checks))
	}
}
