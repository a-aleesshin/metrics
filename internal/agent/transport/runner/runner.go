package runner

import (
	"context"
	"time"
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
}

func NewAgentRunner(
	collectUseCase CollectMetricsExecutor,
	reportUseCase ReportMetricsExecutor,
	pollInterval time.Duration,
	reportInterval time.Duration,
) *AgentRunner {
	return &AgentRunner{
		collectUseCase: collectUseCase,
		reportUseCase:  reportUseCase,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
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
				return err
			}
		case <-tickerSecond.C:
			err := r.collectUseCase.Execute()

			if err != nil {
				return err
			}
		}
	}
}
