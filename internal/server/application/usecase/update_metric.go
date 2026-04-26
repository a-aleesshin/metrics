package usecase

import (
	"strconv"

	"github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"github.com/google/uuid"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type SnapshotSaver interface {
	Execute() error
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

func (u *UpdateMetric) Execute(cmd UpdateMetricCommand) error {
	u.logger.Info("Executing update metric usecase", logger.String("name", cmd.Name))

	name, err := metric.NewName(cmd.Name)
	if err != nil {
		return err
	}

	switch cmd.Type {
	case "gauge":
		return u.updateGauge(name, cmd.Value)
	case "counter":
		return u.updateCounter(name, cmd.Value)
	default:
		return metric.ErrUnsupportedMetricType
	}
}

func (u *UpdateMetric) updateGauge(name metric.Name, rawValue string) error {
	u.logger.Info("Updating gauge metric", logger.String("name", name.String()))

	value, err := strconv.ParseFloat(rawValue, 64)

	if err != nil {
		return metric.ErrInvalidMetricValue
	}

	gauge, err := u.repo.GetGaugeByName(name)

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

	if err := u.repo.SaveGauge(gauge); err != nil {
		return err
	}

	return u.persistSnapshotIfNeeded()
}

func (u *UpdateMetric) updateCounter(name metric.Name, rawValue string) error {
	u.logger.Info("Updating counter metric", logger.String("name", name.String()))

	delta, err := strconv.ParseInt(rawValue, 10, 64)

	if err != nil {
		return metric.ErrInvalidMetricValue
	}

	counter, err := u.repo.GetCounterByName(name)

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

	if err := u.repo.SaveCounter(counter); err != nil {
		return err
	}

	return u.persistSnapshotIfNeeded()
}

func (u *UpdateMetric) persistSnapshotIfNeeded() error {
	if u.snapshotSaver == nil {
		return nil
	}

	if err := u.snapshotSaver.Execute(); err != nil {
		u.logger.Error(
			"snapshot save failed",
			logger.String("component", "update_metric"),
			logger.String("error", err.Error()),
		)

		return nil
	}

	return nil
}
