package main

import (
	"log"
	"net/http"

	inboundhttp "github.com/a-aleesshin/metrics/internal/server/adapters/inbound/http"
	"github.com/a-aleesshin/metrics/internal/server/adapters/inbound/http/metrics"
	"github.com/a-aleesshin/metrics/internal/server/adapters/outbound/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
)

func main() {
	storage := memory.NewMemStorage()
	updateMetrics := usecase.NewUpdateMetric(storage)
	metricsHandler := metrics.NewHandler(updateMetrics)
	router := inboundhttp.NewRouter(metricsHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
