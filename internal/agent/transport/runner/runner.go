package runner

import (
	"context"
	"time"

	portlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

type CollectMetricsExecutor interface {
	Execute() error
}

type ReportMetricsExecutor interface {
	Execute() error
}

type AgentRunner struct {
	collectUseCase CollectMetricsExecutor
	reportUseCase  ReportMetricsExecutor
	pollInterval   time.Duration
	reportInterval time.Duration
	logger         portlogger.Logger
}

func NewAgentRunner(
	collectUseCase CollectMetricsExecutor,
	reportUseCase ReportMetricsExecutor,
	pollInterval time.Duration,
	reportInterval time.Duration,
	logger portlogger.Logger,
) *AgentRunner {
	return &AgentRunner{
		collectUseCase: collectUseCase,
		reportUseCase:  reportUseCase,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		logger:         logger,
	}
}

func (r *AgentRunner) Run(ctx context.Context) error {
	tickerFirst := time.NewTicker(r.reportInterval)
	tickerSecond := time.NewTicker(r.pollInterval)

	defer tickerFirst.Stop()
	defer tickerSecond.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tickerFirst.C:
			err := r.reportUseCase.Execute()

			if err != nil {
				r.logger.Error("report metrics failed", portlogger.Err(err))
				continue
			}
		case <-tickerSecond.C:
			err := r.collectUseCase.Execute()

			if err != nil {
				r.logger.Error("collect metrics failed", portlogger.Err(err))
				continue
			}
		}
	}
}
