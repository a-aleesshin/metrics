package main

import (
	"log"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/metrics"
	sharedrouter "github.com/a-aleesshin/metrics/internal/shared/router"
)

func main() {
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
		Addr:    ":8080",
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
