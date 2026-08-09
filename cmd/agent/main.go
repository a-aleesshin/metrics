package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/a-aleesshin/metrics/internal/agent/application/usecase"
	httpadapter "github.com/a-aleesshin/metrics/internal/agent/infra/http"
	"github.com/a-aleesshin/metrics/internal/agent/infra/persistence/memory"
	randomadapter "github.com/a-aleesshin/metrics/internal/agent/infra/random"
	runtimeadapter "github.com/a-aleesshin/metrics/internal/agent/infra/runtime"
	systemadapter "github.com/a-aleesshin/metrics/internal/agent/infra/system"
	"github.com/a-aleesshin/metrics/internal/agent/transport/cli"
	"github.com/a-aleesshin/metrics/internal/agent/transport/runner"
	"github.com/a-aleesshin/metrics/internal/platform/buildinfo"
	"github.com/a-aleesshin/metrics/internal/platform/logger"
	"go.uber.org/zap"
)

// Информация о сборке, значения подставляются при сборке через
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	buildinfo.Print(buildVersion, buildDate, buildCommit)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run собирает зависимости агента и запускает его до сигнала завершения.
func run() error {
	flags, err := cli.LoadConfig(os.Args[1:])

	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	rider := runtimeadapter.NewMetricRuntimeReader()
	systemReader := systemadapter.NewGopsutilReader()
	repository := memory.NewMemMetricRepository()
	randomValue := randomadapter.NewRandomValueAdapter()

	serverURL := flags.Address
	retryClient := httpadapter.NewRetryClient(http.DefaultClient)
	signingClient := httpadapter.NewSigningClient(retryClient, flags.KeySignature)
	sender := httpadapter.NewMetricSender(serverURL, signingClient)

	collectUsecase := usecase.NewCollectMetricsUseCase(rider, repository, randomValue)
	collectSystemUsecase := usecase.NewCollectSystemMetricsUseCase(systemReader, repository)
	reportUsecase := usecase.NewReportMetricsUseCase(repository, sender)

	baseZap, err := zap.NewProduction()

	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() { _ = baseZap.Sync() }()

	appLogger := logger.NewZapLogger(baseZap)

	agentRunner := runner.NewAgentRunner(
		collectUsecase,
		collectSystemUsecase,
		reportUsecase,
		flags.PollInterval,
		flags.ReportInterval,
		flags.RateLimit,
		appLogger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return agentRunner.Run(ctx)
}
