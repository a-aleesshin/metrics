package usecase

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type UpdateMetricCommand struct {
	Type  string
	Name  string
	Value string
}

type UpdateMetric struct {
	repo repository.MetricRepository
}

func NewUpdateMetric(repo repository.MetricRepository) *UpdateMetric {
	return &UpdateMetric{repo: repo}
}

func (u *UpdateMetric) Execute(cmd UpdateMetricCommand) error {
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

	return u.repo.SaveGauge(gauge)
}

func (u *UpdateMetric) updateCounter(name metric.Name, rawValue string) error {
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

	return u.repo.SaveCounter(counter)
}
