package main

import (
	"log"
	"net/http"
	"os"

	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/metrics"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/middleware"
	"github.com/a-aleesshin/metrics/internal/shared/logger"
	sharedrouter "github.com/a-aleesshin/metrics/internal/shared/router"
	"go.uber.org/zap"
)

func main() {
	flags, err := cli.LoadConfig(os.Args[1:])

	if err != nil {
		log.Fatal("config error: ", err)
	}

	baseZap, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = baseZap.Sync() }()

	storage := memory.NewMemStorage()
	appLogger := logger.NewZapLogger(baseZap)

	updateMetrics := usecase.NewUpdateMetric(storage, appLogger)
	getValueMetric := usecase.NewGetValueMetricUseCase(storage)
	listMetrics := usecase.NewListMetricUseCase(storage)

	metricsHandler := metrics.NewHandler(
		updateMetrics,
		getValueMetric,
		listMetrics,
	)

	router := sharedrouter.New(
		[]func(http.Handler) http.Handler{
			middleware.DecompressRequest,
			middleware.CompressResponse,
			middleware.RequestLogger(baseZap),
		},
		metricsHandler,
	)

	server := &http.Server{
		Addr:    flags.Address,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
