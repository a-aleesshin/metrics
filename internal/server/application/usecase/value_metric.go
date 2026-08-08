package usecase

import (
	"context"
	"strconv"

	applicationerror "github.com/a-aleesshin/metrics/internal/server/application/error"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
)

// ValueMetricCommand — входные данные запроса значения метрики по типу и имени.
type ValueMetricCommand struct {
	Type string
	Name string
}

// GetValueMetricUseCase — use case чтения текущего значения метрики.
type GetValueMetricUseCase struct {
	repo repository.MetricQueryRepository
}

// NewGetValueMetricUseCase создаёт use case чтения значения метрики.
func NewGetValueMetricUseCase(repo repository.MetricQueryRepository) *GetValueMetricUseCase {
	return &GetValueMetricUseCase{repo: repo}
}

// Execute возвращает значение метрики строкой; если метрика не найдена —
// applicationerror.ErrMetricNotFound, для неизвестного типа — metric.ErrUnsupportedMetricType.
func (u *GetValueMetricUseCase) Execute(ctx context.Context, cmd ValueMetricCommand) (string, error) {
	metricName := metric.Name(cmd.Name)

	switch cmd.Type {
	case "gauge":
		data, found, err := u.repo.FindGaugeByName(ctx, metricName)

		if err != nil {
			return "", err
		}

		if !found {
			return "", applicationerror.ErrMetricNotFound
		}

		return strconv.FormatFloat(data, 'f', -1, 64), nil
	case "counter":
		data, found, err := u.repo.FindCounterByName(ctx, metricName)
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
