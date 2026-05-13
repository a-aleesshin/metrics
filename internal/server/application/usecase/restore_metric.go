package usecase

import (
	"context"
	"fmt"

	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
)

type RestoreMetricUseCase struct {
	repository         repository.MetricRepository
	repositorySnapshot repository.MetricSnapshotStore
	mapper             *mapper.MetricSnapshotMapper
}

func NewRestoreMetricUseCase(repository repository.MetricRepository, repositorySnapshot repository.MetricSnapshotStore, mapper *mapper.MetricSnapshotMapper) *RestoreMetricUseCase {
	return &RestoreMetricUseCase{
		repository:         repository,
		repositorySnapshot: repositorySnapshot,
		mapper:             mapper,
	}
}

func (u *RestoreMetricUseCase) Execute(ctx context.Context) error {
	snapshots, err := u.repositorySnapshot.Load(ctx)

	if err != nil {
		return fmt.Errorf("load snapshots: %w", err)
	}

	for i, snapshot := range snapshots {
		gauge, counter, err := u.mapper.SnapshotToDomain(snapshot)

		if err != nil {
			return fmt.Errorf("convert snapshot %d to domain: %w", i, err)
		}

		if gauge == nil && counter == nil {
			return fmt.Errorf("snapshot %d has no domain object", i)
		}

		if gauge != nil {
			if err := u.repository.SaveGauge(ctx, gauge); err != nil {
				return fmt.Errorf("save gauge %d: %w", i, err)
			}

			continue
		}

		if err := u.repository.SaveCounter(ctx, counter); err != nil {
			return fmt.Errorf("save counter %d: %w", i, err)
		}
	}

	return nil
}
