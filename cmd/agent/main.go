package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/a-aleesshin/metrics/internal/agent/application/usecase"
	"github.com/a-aleesshin/metrics/internal/agent/infra/http"
	"github.com/a-aleesshin/metrics/internal/agent/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/agent/infra/random"
	"github.com/a-aleesshin/metrics/internal/agent/infra/runtime"
	"github.com/a-aleesshin/metrics/internal/agent/transport/runner"
)

func main() {
	rider := runtimeadapter.NewMetricRuntimeReader()
	repository := memory.NewMemMetricRepository()
	randomValue := randomadapter.NewRandomValueAdapter()

	serverUrl := "http://localhost:8080"
	sender := httpadapter.NewMetricSender(serverUrl, http.DefaultClient)

	collectUsecase := usecase.NewCollectMetricsUseCase(rider, repository, randomValue)
	reportUsecase := usecase.NewReportMetricsUseCase(repository, sender)

	agentRunner := runner.NewAgentRunner(
		collectUsecase,
		reportUsecase,
		2,
		10,
	)

	err := agentRunner.Run()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
