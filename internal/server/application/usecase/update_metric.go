package usecase

import (
	"context"
	"strconv"

	"github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"github.com/google/uuid"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type SnapshotSaver interface {
	Execute(ctx context.Context) error
}

type UpdateMetricCommand struct {
	Type  string
	Name  string
	Value string
}

type UpdateMetric struct {
	repo          repository.MetricRepository
	logger        logger.Logger
	snapshotSaver SnapshotSaver
}

func NewUpdateMetric(repo repository.MetricRepository, logger logger.Logger, snapshotSaver SnapshotSaver) *UpdateMetric {
	return &UpdateMetric{
		repo:          repo,
		logger:        logger,
		snapshotSaver: snapshotSaver,
	}
}

func (u *UpdateMetric) Execute(ctx context.Context, cmd UpdateMetricCommand) error {
	u.logger.Info("Executing update metric usecase", logger.String("name", cmd.Name))

	name, err := metric.NewName(cmd.Name)
	if err != nil {
		return err
	}

	switch cmd.Type {
	case "gauge":
		return u.updateGauge(ctx, name, cmd.Value)
	case "counter":
		return u.updateCounter(ctx, name, cmd.Value)
	default:
		return metric.ErrUnsupportedMetricType
	}
}

func (u *UpdateMetric) updateGauge(ctx context.Context, name metric.Name, rawValue string) error {
	u.logger.Info("Updating gauge metric", logger.String("name", name.String()))

	value, err := strconv.ParseFloat(rawValue, 64)

	if err != nil {
		return metric.ErrInvalidMetricValue
	}

	gauge, err := u.repo.GetGaugeByName(ctx, name)

	if err != nil {
		return err
	}

	if gauge == nil {
		gauge, err = metric.NewGauge(uuid.NewString(), name.String(), value)

		if err != nil {
			return err
		}
	} else {
		gauge.UpdateValue(value)
	}

	if err := u.repo.SaveGauge(ctx, gauge); err != nil {
		return err
	}

	return u.persistSnapshotIfNeeded(ctx)
}

func (u *UpdateMetric) updateCounter(ctx context.Context, name metric.Name, rawValue string) error {
	u.logger.Info("Updating counter metric", logger.String("name", name.String()))

	delta, err := strconv.ParseInt(rawValue, 10, 64)

	if err != nil {
		return metric.ErrInvalidMetricValue
	}

	counter, err := u.repo.GetCounterByName(ctx, name)

	if err != nil {
		return err
	}

	if counter == nil {
		counter, err = metric.NewCounter(uuid.NewString(), name.String(), delta)

		if err != nil {
			return err
		}
	} else {
		counter.Add(delta)
	}

	if err := u.repo.SaveCounter(ctx, counter); err != nil {
		return err
	}

	return u.persistSnapshotIfNeeded(ctx)
}

func (u *UpdateMetric) persistSnapshotIfNeeded(ctx context.Context) error {
	if u.snapshotSaver == nil {
		return nil
	}

	if err := u.snapshotSaver.Execute(ctx); err != nil {
		u.logger.Error(
			"snapshot save failed",
			logger.String("component", "update_metric"),
			logger.String("error", err.Error()),
		)

		return nil
	}

	return nil
}
