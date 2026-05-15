package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type batchRepositorySpy struct {
	batch  repository.MetricBatch
	err    error
	called bool
}

func (s *batchRepositorySpy) UpdateBatch(ctx context.Context, batch repository.MetricBatch) error {
	s.called = true
	s.batch = batch

	if s.err != nil {
		return s.err
	}

	return nil
}

type idGeneratorStub struct {
	ids []metric.ID
	err error
	n   int
}

func (s *idGeneratorStub) NewID() (metric.ID, error) {
	if s.err != nil {
		return "", s.err
	}

	if s.n >= len(s.ids) {
		return metric.ID("generated-id"), nil
	}

	id := s.ids[s.n]
	s.n++

	return id, nil
}

func TestUpdatesMetricsUseCase_Execute_UpdatesBatch(t *testing.T) {
	// Arrange
	repo := &batchRepositorySpy{}
	idGen := &idGeneratorStub{
		ids: []metric.ID{"gauge-id", "counter-id"},
	}
	uc := NewUpdatesMetricsUseCase(repo, idGen)

	command := UpdatesMetricsCommand{
		Metrics: []MetricUpdatesCommand{
			{
				Name:  "Alloc",
				MType: "gauge",
				Value: float64Ptr(123.45),
			},
			{
				Name:  "PollCount",
				MType: "counter",
				Delta: int64Ptr(7),
			},
		},
	}

	// Act
	err := uc.Execute(t.Context(), command)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.called {
		t.Fatal("expected repository to be called")
	}

	if len(repo.batch.Gauges) != 1 {
		t.Fatalf("expected 1 gauge, got %d", len(repo.batch.Gauges))
	}

	gauge := repo.batch.Gauges[0]
	if gauge.Id().String() != "gauge-id" {
		t.Fatalf("expected gauge id %q, got %q", "gauge-id", gauge.Id().String())
	}
	if gauge.Name().String() != "Alloc" {
		t.Fatalf("expected gauge name Alloc, got %s", gauge.Name().String())
	}
	if gauge.Value() != 123.45 {
		t.Fatalf("expected gauge value 123.45, got %v", gauge.Value())
	}

	if len(repo.batch.Counters) != 1 {
		t.Fatalf("expected 1 counter, got %d", len(repo.batch.Counters))
	}

	counter := repo.batch.Counters[0]
	if counter.Id().String() != "counter-id" {
		t.Fatalf("expected counter id %q, got %q", "counter-id", counter.Id().String())
	}
	if counter.Name().String() != "PollCount" {
		t.Fatalf("expected counter name PollCount, got %s", counter.Name().String())
	}
	if counter.Delta() != 7 {
		t.Fatalf("expected counter delta 7, got %d", counter.Delta())
	}
}

func TestUpdatesMetricsUseCase_Execute_EmptyBatchDoesNotCallRepository(t *testing.T) {
	// Arrange
	repo := &batchRepositorySpy{}
	idGen := &idGeneratorStub{}
	uc := NewUpdatesMetricsUseCase(repo, idGen)

	// Act
	err := uc.Execute(t.Context(), UpdatesMetricsCommand{})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.called {
		t.Fatal("expected repository not to be called")
	}
}

func TestUpdatesMetricsUseCase_Execute_ReturnsError(t *testing.T) {
	repoErr := errors.New("repo failed")
	idErr := errors.New("id failed")

	tests := []struct {
		name    string
		command UpdatesMetricsCommand
		repo    *batchRepositorySpy
		idGen   *idGeneratorStub
		wantErr error
	}{
		{
			name: "id generator error",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge", Value: float64Ptr(1)},
				},
			},
			repo:    &batchRepositorySpy{},
			idGen:   &idGeneratorStub{err: idErr},
			wantErr: idErr,
		},
		{
			name: "repository error",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge", Value: float64Ptr(1)},
				},
			},
			repo:    &batchRepositorySpy{err: repoErr},
			idGen:   &idGeneratorStub{ids: []metric.ID{"id-1"}},
			wantErr: repoErr,
		},
		{
			name: "empty name",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "", MType: "gauge", Value: float64Ptr(1)},
				},
			},
			repo:    &batchRepositorySpy{},
			idGen:   &idGeneratorStub{ids: []metric.ID{"id-1"}},
			wantErr: metric.ErrNameEmpty,
		},
		{
			name: "unsupported type",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "Alloc", MType: "histogram", Value: float64Ptr(1)},
				},
			},
			repo:    &batchRepositorySpy{},
			idGen:   &idGeneratorStub{ids: []metric.ID{"id-1"}},
			wantErr: metric.ErrUnsupportedMetricType,
		},
		{
			name: "gauge value is nil",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "Alloc", MType: "gauge"},
				},
			},
			repo:    &batchRepositorySpy{},
			idGen:   &idGeneratorStub{ids: []metric.ID{"id-1"}},
			wantErr: metric.ErrInvalidMetricValue,
		},
		{
			name: "counter delta is nil",
			command: UpdatesMetricsCommand{
				Metrics: []MetricUpdatesCommand{
					{Name: "PollCount", MType: "counter"},
				},
			},
			repo:    &batchRepositorySpy{},
			idGen:   &idGeneratorStub{ids: []metric.ID{"id-1"}},
			wantErr: metric.ErrInvalidMetricValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := NewUpdatesMetricsUseCase(tt.repo, tt.idGen)

			// Act
			err := uc.Execute(t.Context(), tt.command)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
