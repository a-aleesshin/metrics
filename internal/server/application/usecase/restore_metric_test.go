package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type metricRepositoryStubREUC struct {
	saveGaugeErr   error
	saveCounterErr error

	savedGauges   []*metric.Gauge
	savedCounters []*metric.Counter
}

func (m *metricRepositoryStubREUC) GetGaugeByName(name metric.Name) (*metric.Gauge, error) {
	return nil, nil
}

func (m *metricRepositoryStubREUC) GetCounterByName(name metric.Name) (*metric.Counter, error) {
	return nil, nil
}

func (m *metricRepositoryStubREUC) SaveGauge(g *metric.Gauge) error {
	if m.saveGaugeErr != nil {
		return m.saveGaugeErr
	}
	m.savedGauges = append(m.savedGauges, g)
	return nil
}

func (m *metricRepositoryStubREUC) SaveCounter(c *metric.Counter) error {
	if m.saveCounterErr != nil {
		return m.saveCounterErr
	}
	m.savedCounters = append(m.savedCounters, c)
	return nil
}

type metricSnapshotStoreStub struct {
	toLoad  []repository.MetricSnapshot
	loadErr error
}

func (s *metricSnapshotStoreStub) Save(metrics []repository.MetricSnapshot) error {
	return nil
}

func (s *metricSnapshotStoreStub) Load() ([]repository.MetricSnapshot, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return append([]repository.MetricSnapshot(nil), s.toLoad...), nil
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestRestoreMetricUseCase_Execute(t *testing.T) {
	tests := []struct {
		name             string
		toLoad           []repository.MetricSnapshot
		loadErr          error
		saveGaugeErr     error
		saveCounterErr   error
		wantErrContains  string
		wantGaugeSaves   int
		wantCounterSaves int
	}{
		{
			name: "user_case_empty_snapshots",
		},
		{
			name: "user_case_gauge_snapshot_success",
			toLoad: []repository.MetricSnapshot{
				{
					ID:    "PollCount",
					Type:  "counter",
					Delta: int64Ptr(7),
				},
			},
			wantCounterSaves: 1,
		},
		{
			name: "user_case_counter_snapshot_success",
			toLoad: []repository.MetricSnapshot{
				{
					ID:    "PollCount",
					Type:  "counter",
					Delta: int64Ptr(7),
				},
			},
			wantCounterSaves: 1,
		},
		{
			name:            "user_case_load_error",
			loadErr:         errors.New("load failed"),
			wantErrContains: "load snapshots: load failed",
		},
		{
			name: "user_case_map_error",
			toLoad: []repository.MetricSnapshot{
				{
					ID:   "x",
					Type: "unknown",
				},
			},
			wantErrContains: "convert snapshot 0 to domain",
		},
		{
			name: "user_case_save_gauge_error",
			toLoad: []repository.MetricSnapshot{
				{
					ID:    "Alloc",
					Type:  "gauge",
					Value: float64Ptr(1),
				},
			},
			saveGaugeErr:    errors.New("save gauge failed"),
			wantErrContains: "save gauge 0",
		},
		{
			name: "save counter error",
			toLoad: []repository.MetricSnapshot{
				{
					ID:    "PollCount",
					Type:  "counter",
					Delta: int64Ptr(1),
				},
			},
			saveCounterErr:  errors.New("save counter failed"),
			wantErrContains: "save counter 0",
		},
	}

	// Arrange
	snapshotMapper := mapper.NewMetricSnapshotMapper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &metricRepositoryStubREUC{
				saveGaugeErr:   tt.saveGaugeErr,
				saveCounterErr: tt.saveCounterErr,
			}

			store := &metricSnapshotStoreStub{
				toLoad:  tt.toLoad,
				loadErr: tt.loadErr,
			}

			uc := NewRestoreMetricUseCase(repo, store, snapshotMapper)

			// Act
			err := uc.Execute()

			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected error %q, got %q", tt.wantErrContains, err.Error())
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(repo.savedGauges) != tt.wantGaugeSaves {
				t.Fatalf("expected %d gauge saves, got %d", tt.wantGaugeSaves, len(repo.savedGauges))
			}

			if len(repo.savedCounters) != tt.wantCounterSaves {
				t.Fatalf("expected %d counter saves, got %d", tt.wantCounterSaves, len(repo.savedCounters))
			}
		})
	}
}
