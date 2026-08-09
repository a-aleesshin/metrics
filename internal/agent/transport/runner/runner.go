// Package runner запускает жизненный цикл агента: периодический сбор метрик
// и их отправку пулом воркеров.
package runner

import (
	"context"
	"sync"
	"time"

	"github.com/a-aleesshin/metrics/internal/agent/application/dto"
	portlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

// CollectMetricsExecutor — сценарий одного цикла сбора метрик.
type CollectMetricsExecutor interface {
	// Execute выполняет один цикл сбора метрик.
	Execute() error
}

// ReportMetricsExecutor — сценарий формирования и отправки отчёта о метриках.
type ReportMetricsExecutor interface {
	// BuildMetrics формирует список метрик для отправки.
	BuildMetrics() ([]dto.MetricDTO, error)
	// SendMetrics отправляет подготовленный список метрик.
	SendMetrics(metrics []dto.MetricDTO) error
}

// AgentRunner координирует горутины сбора и отправки метрик по расписанию.
type AgentRunner struct {
	collectUseCase       CollectMetricsExecutor
	collectSystemUseCase CollectMetricsExecutor
	reportUseCase        ReportMetricsExecutor
	pollInterval         time.Duration
	reportInterval       time.Duration
	rateLimit            int
	logger               portlogger.Logger
}

// NewAgentRunner создаёт раннер агента; rateLimit <= 0 приводится к 1.
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

// Run запускает сборщики и rateLimit воркеров отправки; блокируется до отмены ctx.
func (r *AgentRunner) Run(ctx context.Context) error {
	jobs := make(chan []dto.MetricDTO)
	var wg sync.WaitGroup

	for i := 0; i < r.rateLimit; i++ {
		wg.Add(1)
		go r.runReportWorker(&wg, jobs)
	}

	wg.Add(1)
	go r.runCollector(ctx, &wg)

	if r.collectSystemUseCase != nil {
		wg.Add(1)
		go r.runSystemCollector(ctx, &wg)
	}

	reportTicker := time.NewTicker(r.reportInterval)
	defer reportTicker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
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
				break loop
			}
		}
	}

	close(jobs)
	wg.Wait()

	r.reportFinal()

	return nil
}

// reportFinal синхронно отправляет текущее состояние метрик; ошибки логируются,
// но не прерывают завершение.
func (r *AgentRunner) reportFinal() {
	metrics, err := r.reportUseCase.BuildMetrics()
	if err != nil {
		r.logger.Error("build final metrics report failed", portlogger.Err(err))
		return
	}

	if len(metrics) == 0 {
		return
	}

	if err := r.reportUseCase.SendMetrics(metrics); err != nil {
		r.logger.Error("send final metrics report failed", portlogger.Err(err))
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

// runReportWorker отправляет батчи из очереди до её закрытия. Воркер намеренно
// не следит за ctx: при завершении он дорабатывает уже принятые батчи.
func (r *AgentRunner) runReportWorker(wg *sync.WaitGroup, jobs <-chan []dto.MetricDTO) {
	defer wg.Done()

	for metrics := range jobs {
		if err := r.reportUseCase.SendMetrics(metrics); err != nil {
			r.logger.Error("report metrics failed", portlogger.Err(err))
		}
	}
}
