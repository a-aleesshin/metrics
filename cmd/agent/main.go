package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/a-aleesshin/metrics/internal/agent/application/usecase"
	"github.com/a-aleesshin/metrics/internal/agent/infra/http"
	"github.com/a-aleesshin/metrics/internal/agent/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/agent/infra/random"
	"github.com/a-aleesshin/metrics/internal/agent/infra/runtime"
	"github.com/a-aleesshin/metrics/internal/agent/transport/cli"
	"github.com/a-aleesshin/metrics/internal/agent/transport/runner"
)

func main() {
	flags, err := cli.ParseConfig(os.Args[1:])

	if err != nil {
		log.Fatal("config error: ", err)
	}

	rider := runtimeadapter.NewMetricRuntimeReader()
	repository := memory.NewMemMetricRepository()
	randomValue := randomadapter.NewRandomValueAdapter()

	serverUrl := flags.Address
	sender := httpadapter.NewMetricSender(serverUrl, http.DefaultClient)

	collectUsecase := usecase.NewCollectMetricsUseCase(rider, repository, randomValue)
	reportUsecase := usecase.NewReportMetricsUseCase(repository, sender)

	agentRunner := runner.NewAgentRunner(
		collectUsecase,
		reportUsecase,
		flags.PollInterval,
		flags.ReportInterval,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = agentRunner.Run(ctx)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
