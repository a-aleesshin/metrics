package usecase

import (
	"strconv"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

type ValueMetricCommand struct {
	Type string
	Name string
}

type GetValueMetricUseCase struct {
	repo repository.MetricQueryRepository
}

func NewGetValueMetricUseCase(repo repository.MetricQueryRepository) *GetValueMetricUseCase {
	return &GetValueMetricUseCase{repo: repo}
}

func (u *GetValueMetricUseCase) Execute(cmd ValueMetricCommand) (string, error) {
	metricName := metric.Name(cmd.Name)

	switch cmd.Type {
	case "gauge":
		data, found, err := u.repo.FindGaugeByName(metricName)

		if err != nil {
			return "", err
		}

		if !found {
			return "", applicationerror.ErrMetricNotFound
		}

		return strconv.FormatFloat(data, 'f', -1, 64), nil
	case "counter":
		data, found, err := u.repo.FindCounterByName(metricName)
		if err != nil {
			return "", err
		}

		if !found {
			return "", applicationerror.ErrMetricNotFound
		}

		return strconv.FormatInt(data, 10), nil
	default:
		return "", metric.ErrUnsupportedMetricType
	}
}
