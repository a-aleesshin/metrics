package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	snapshotfile "github.com/a-aleesshin/metrics/internal/server/infra/persistence/file"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/metrics"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/middleware"
	"github.com/a-aleesshin/metrics/internal/shared/logger"
	sharedrouter "github.com/a-aleesshin/metrics/internal/shared/router"
	"go.uber.org/zap"
)

func run(cfg *cli.ServerConfig) error {
	baseZap, err := zap.NewProduction()

	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = baseZap.Sync() }()

	storage := memory.NewMemStorage()

	snapshotStore, err := snapshotfile.NewSnapshotStore(cfg.FileStoragePath)

	if err != nil {
		return fmt.Errorf("create snapshot store: %w", err)
	}

	appLogger := logger.NewZapLogger(baseZap)
	snapshotMapper := mapper.NewMetricSnapshotMapper()

	saveSnapshotUC := usecase.NewSaveMetricSnapshotUseCase(storage, snapshotStore, snapshotMapper)
	restoreUC := usecase.NewRestoreMetricUseCase(storage, snapshotStore, snapshotMapper)

	if cfg.Restore {
		if err := restoreUC.Execute(); err != nil {
			return fmt.Errorf("restore metrics: %w", err)
		}
	}

	var saver usecase.SnapshotSaver
	if cfg.StoreInterval == 0 {
		saver = saveSnapshotUC
	}

	updateMetricsUC := usecase.NewUpdateMetric(storage, appLogger, saver)
	getValueMetricUC := usecase.NewGetValueMetricUseCase(storage)
	listMetricsUC := usecase.NewListMetricUseCase(storage)

	metricsHandler := metrics.NewHandler(updateMetricsUC, getValueMetricUC, listMetricsUC)

	router := sharedrouter.New(
		[]func(http.Handler) http.Handler{
			middleware.DecompressRequest,
			middleware.CompressResponse,
			middleware.RequestLogger(baseZap),
		},
		metricsHandler,
	)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(cfg.StoreInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := saveSnapshotUC.Execute(); err != nil {
						log.Printf("periodic snapshot save failed: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		if err := saveSnapshotUC.Execute(); err != nil {
			return fmt.Errorf("final snapshot save: %w", err)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server close: %w", err)
		}
	}

	return nil
}
