package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type metricStateRepositoryStub struct {
	state repository.MetricsState
	err   error
}

func (s *metricStateRepositoryStub) GetAllMetrics() (repository.MetricsState, error) {
	if s.err != nil {
		return repository.MetricsState{}, s.err
	}
	return s.state, nil
}

type metricSnapshotStoreStubSaveUC struct {
	saved   []repository.MetricSnapshot
	saveErr error
	loadErr error
}

func (s *metricSnapshotStoreStubSaveUC) Save(metrics []repository.MetricSnapshot) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = append([]repository.MetricSnapshot(nil), metrics...)
	return nil
}

func (s *metricSnapshotStoreStubSaveUC) Load() ([]repository.MetricSnapshot, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return nil, nil
}

func mustGauge(t *testing.T, id, name string, value float64) *metric.Gauge {
	t.Helper()

	g, err := metric.RestoreGauge(id, name, value)

	if err != nil {
		t.Fatalf("restore gauge: %v", err)
	}

	return g
}

func mustCounter(t *testing.T, id, name string, delta int64) *metric.Counter {
	t.Helper()

	c, err := metric.RestoreCounter(id, name, delta)

	if err != nil {
		t.Fatalf("restore counter: %v", err)
	}

	return c
}

func TestSaveMetricSnapshotUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		state         repository.MetricsState
		getStateErr   error
		saveErr       error
		wantErr       string
		wantSaveCalls int
		wantSavedLen  int
	}{
		{
			name:        "user_case_state_repo_error",
			getStateErr: errors.New("state failed"),
			wantErr:     "get all metrics",
		},
		{
			name:          "user_case_empty_state_saved_as_empty_snapshot",
			state:         repository.MetricsState{},
			wantSaveCalls: 1,
			wantSavedLen:  0,
		},
		{
			name: "use_case_success_gauges_and_counters",
			state: repository.MetricsState{
				Gauges: []*metric.Gauge{
					mustGauge(t, "g1", "Alloc", 12.5),
				},
				Counters: []*metric.Counter{
					mustCounter(t, "c1", "PollCount", 7),
				},
			},
			wantSaveCalls: 1,
			wantSavedLen:  2,
		},
		{
			name: "use_case_save_error",
			state: repository.MetricsState{
				Gauges: []*metric.Gauge{
					mustGauge(t, "g1", "Alloc", 1),
				},
			},
			saveErr: errors.New("save failed"),
			wantErr: "error save snapshot",
		},
		{
			name: "use_case_mapping_error_from_nil_gauge",
			state: repository.MetricsState{
				Gauges: []*metric.Gauge{nil},
			},
			wantErr: "convert gauge to snapshot",
		},
		{
			name: "use_case_mapping_error_from_nil_counter",
			state: repository.MetricsState{
				Counters: []*metric.Counter{nil},
			},
			wantErr: "convert counter to snapshot",
		},
	}

	// Arrange
	snapshotMapper := mapper.NewMetricSnapshotMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			stateRepo := &metricStateRepositoryStub{
				state: tt.state,
				err:   tt.getStateErr,
			}

			snapshotStore := &metricSnapshotStoreStubSaveUC{
				saveErr: tt.saveErr,
			}

			uc := NewSaveMetricSnapshotUseCase(stateRepo, snapshotStore, snapshotMapper)

			// Act
			err := uc.Execute()

			// Assert
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}

				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantSaveCalls == 0 && snapshotStore.saved != nil {
				t.Fatalf("expected no saved snapshots, got %+v", snapshotStore.saved)
			}

			if tt.wantSaveCalls > 0 && snapshotStore == nil {
				t.Fatalf("expected saved snapshots, got nil")
			}

			if len(snapshotStore.saved) != tt.wantSavedLen {
				t.Fatalf("expected %d saved snapshots, got %d", tt.wantSavedLen, len(snapshotStore.saved))
			}
		})
	}
}
