package main

import (
	"log"
	"net/http"
	"os"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/metrics"
	sharedrouter "github.com/a-aleesshin/metrics/internal/shared/router"
)

func main() {
	flags, err := cli.ParseConfig(os.Args[1:])

	if err != nil {
		log.Fatal("config error: ", err)
	}

	storage := memory.NewMemStorage()

	updateMetrics := usecase.NewUpdateMetric(storage)
	getValueMetric := usecase.NewGetValueMetricUseCase(storage)
	listMetrics := usecase.NewListMetricUseCase(storage)

	metricsHandler := metrics.NewHandler(
		updateMetrics,
		getValueMetric,
		listMetrics,
	)

	router := sharedrouter.New(metricsHandler)

	server := &http.Server{
		Addr:    flags.Address,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
