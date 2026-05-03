package usecase

import (
	"fmt"
	"sort"

	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

type SaveMetricSnapshotUseCase struct {
	repositoryState    repository.MetricStateRepository
	repositorySnapshot repository.MetricSnapshotStore
	mapper             *mapper.MetricSnapshotMapper
}

func NewSaveMetricSnapshotUseCase(repositoryState repository.MetricStateRepository, repositorySnapshot repository.MetricSnapshotStore, mapper *mapper.MetricSnapshotMapper) *SaveMetricSnapshotUseCase {
	return &SaveMetricSnapshotUseCase{
		repositoryState:    repositoryState,
		repositorySnapshot: repositorySnapshot,
		mapper:             mapper,
	}
}

func (u *SaveMetricSnapshotUseCase) Execute() error {
	metrics, err := u.repositoryState.GetAllMetrics()

	if err != nil {
		return fmt.Errorf("get all metrics: %w", err)
	}

	snapshots := make([]repository.MetricSnapshot, 0, len(metrics.Counters)+len(metrics.Gauges))

	for _, counter := range metrics.Counters {
		cs, err := u.mapper.CounterToSnapshot(counter)

		if err != nil {
			return fmt.Errorf("convert counter to snapshot: %w", err)
		}

		snapshots = append(snapshots, cs)
	}

	for _, gauge := range metrics.Gauges {
		gs, err := u.mapper.GaugeToSnapshot(gauge)

		if err != nil {
			return fmt.Errorf("convert gauge to snapshot: %w", err)
		}

		snapshots = append(snapshots, gs)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].Type == snapshots[j].Type {
			return snapshots[i].ID < snapshots[j].ID
		}
		return snapshots[i].Type < snapshots[j].Type
	})

	if err := u.repositorySnapshot.Save(snapshots); err != nil {
		return fmt.Errorf("error save snapshot: %w", err)
	}

	return nil
}
