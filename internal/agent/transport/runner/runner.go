package runner

import (
	"context"
	"sync"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	portlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

type CollectMetricsExecutor interface {
	Execute() error
}

type ReportMetricsExecutor interface {
	BuildMetrics() ([]dto.MetricDTO, error)
	SendMetrics(metrics []dto.MetricDTO) error
}

type AgentRunner struct {
	collectUseCase       CollectMetricsExecutor
	collectSystemUseCase CollectMetricsExecutor
	reportUseCase        ReportMetricsExecutor
	pollInterval         time.Duration
	reportInterval       time.Duration
	rateLimit            int
	logger               portlogger.Logger
}

func NewAgentRunner(
	collectUseCase CollectMetricsExecutor,
	collectSystemUseCase CollectMetricsExecutor,
	reportUseCase ReportMetricsExecutor,
	pollInterval time.Duration,
	reportInterval time.Duration,
	rateLimit int,
	logger portlogger.Logger,
) *AgentRunner {
	if rateLimit <= 0 {
		rateLimit = 1
	}

	return &AgentRunner{
		collectUseCase:       collectUseCase,
		collectSystemUseCase: collectSystemUseCase,
		reportUseCase:        reportUseCase,
		pollInterval:         pollInterval,
		reportInterval:       reportInterval,
		rateLimit:            rateLimit,
		logger:               logger,
	}
}

func (r *AgentRunner) Run(ctx context.Context) error {
	jobs := make(chan []dto.MetricDTO)
	var wg sync.WaitGroup

	for i := 0; i < r.rateLimit; i++ {
		wg.Add(1)
		go r.runReportWorker(ctx, &wg, jobs)
	}

	wg.Add(1)
	go r.runCollector(ctx, &wg)

	if r.collectSystemUseCase != nil {
		wg.Add(1)
		go r.runSystemCollector(ctx, &wg)
	}

	reportTicker := time.NewTicker(r.reportInterval)
	defer reportTicker.Stop()
	defer func() {
		close(jobs)
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reportTicker.C:
			metrics, err := r.reportUseCase.BuildMetrics()
			if err != nil {
				r.logger.Error("build metrics report failed", portlogger.Err(err))
				continue
			}

			if len(metrics) == 0 {
				continue
			}

			select {
			case jobs <- metrics:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (r *AgentRunner) runCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.collectUseCase.Execute(); err != nil {
				r.logger.Error("collect metrics failed", portlogger.Err(err))
			}
		}
	}
}

func (r *AgentRunner) runSystemCollector(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.collectSystemUseCase.Execute(); err != nil {
				r.logger.Error("collect system metrics failed", portlogger.Err(err))
			}
		}
	}
}

func (r *AgentRunner) runReportWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan []dto.MetricDTO) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case metrics, ok := <-jobs:
			if !ok {
				return
			}

			if err := r.reportUseCase.SendMetrics(metrics); err != nil {
				r.logger.Error("report metrics failed", portlogger.Err(err))
			}
		}
	}
}
