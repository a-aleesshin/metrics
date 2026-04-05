package metrics

import (
	"github.com/a-aleesshin/metrics/internal/server/application/dto"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
)

type updateUseCaseNoop struct{}

func (updateUseCaseNoop) Execute(command usecase.UpdateMetricCommand) error { return nil }

type listUseCaseNoop struct{}

func (listUseCaseNoop) Execute() (dto.ListMetricsResult, error) { return dto.ListMetricsResult{}, nil }

type valueUseCaseNoop struct{}

func (valueUseCaseNoop) Execute(cmd usecase.ValueMetricCommand) (string, error) { return "", nil }
