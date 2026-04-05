package runner

import (
	"time"
)

// TODO протестирую позже, когда вынесу time.NewTicker и создам фабрику которую будет получать runner

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

func (r *AgentRunner) Run() error {
	tickerFirst := time.NewTicker(r.reportInterval * time.Second)
	tickerSecond := time.NewTicker(r.pollInterval * time.Second)

	defer tickerFirst.Stop()
	defer tickerSecond.Stop()

	for {
		select {
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
