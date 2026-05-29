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

	platformpostgres "github.com/a-aleesshin/metrics/internal/platform/db/postgres"
	"github.com/a-aleesshin/metrics/internal/platform/health"
	sharedrouter "github.com/a-aleesshin/metrics/internal/platform/http"
	"github.com/a-aleesshin/metrics/internal/platform/id"
	platformlogger "github.com/a-aleesshin/metrics/internal/platform/logger"
	"github.com/a-aleesshin/metrics/internal/server/application/mapper"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/application/usecase"
	snapshotfile "github.com/a-aleesshin/metrics/internal/server/infra/persistence/file"
	"github.com/a-aleesshin/metrics/internal/server/infra/persistence/memory"
	storagepostgres "github.com/a-aleesshin/metrics/internal/server/infra/persistence/postgres"
	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/healths"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics"
	"github.com/a-aleesshin/metrics/internal/server/transport/http/middleware"
	sharedlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"go.uber.org/zap"
)

type storageRuntime struct {
	metricRepo    repository.MetricRepository
	queryRepo     repository.MetricQueryRepository
	healthService *health.Service
	snapshotSaver usecase.SnapshotSaver
	periodicSaver usecase.SnapshotSaver
	batchRepo     repository.MetricBatchRepository
	cleanup       func()
}

type appLoggerRuntime struct {
	httpLogger *zap.Logger
	appLogger  sharedlogger.Logger
	cleanup    func()
}

func run(cfg *cli.ServerConfig) error {
	startupCtx := context.Background()

	loggers, err := buildLoggers()
	if err != nil {
		return err
	}
	defer loggers.cleanup()

	runtime, err := buildStorageRuntime(startupCtx, cfg)
	if err != nil {
		return err
	}
	defer runtime.cleanup()

	updateMetricsUC := usecase.NewUpdateMetric(
		runtime.metricRepo,
		loggers.appLogger,
		runtime.snapshotSaver,
	)
	getValueMetricUC := usecase.NewGetValueMetricUseCase(runtime.queryRepo)
	listMetricsUC := usecase.NewListMetricUseCase(runtime.queryRepo)

	// TODO спорный случай, на данный момент uuid сущности не нужен, но я пока его оставлю
	idGenerator := id.NewUUIDV7Generator()

	updatesMetricsUC := usecase.NewUpdatesMetricsUseCase(
		runtime.batchRepo,
		idGenerator,
	)

	updateHandler := metrics.NewUpdateHandler(updateMetricsUC)
	updateJSONHandler := metrics.NewUpdateJsonHandler(updateMetricsUC)
	updatesHandler := metrics.NewUpdatesHandler(updatesMetricsUC)

	valueHandler := metrics.NewValueHandler(getValueMetricUC)
	valueJSONHandler := metrics.NewValueJsonHandler(getValueMetricUC)

	listHandler := metrics.NewListMetricsHandler(listMetricsUC)

	metricsHandler := metrics.NewHandler(
		updateHandler,
		updateJSONHandler,
		updatesHandler,
		valueHandler,
		valueJSONHandler,
		listHandler,
	)

	pingHandler := healths.NewPingHandler(runtime.healthService)
	healthHandler := healths.NewHandler(pingHandler)

	router := sharedrouter.New(
		[]func(http.Handler) http.Handler{
			middleware.WithHashSHA256(cfg.KeySignature),
			middleware.DecompressRequest,
			middleware.CompressResponse,
			middleware.RequestLogger(loggers.httpLogger),
		},
		metricsHandler,
		healthHandler,
	)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startPeriodicSnapshot(serverCtx, cfg.StoreInterval, runtime.periodicSaver)

	return serve(serverCtx, server, runtime.periodicSaver)
}

func startPeriodicSnapshot(ctx context.Context, interval time.Duration, saver usecase.SnapshotSaver) {
	if saver == nil || interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := saver.Execute(ctx); err != nil {
					log.Printf("periodic snapshot save failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func serve(ctx context.Context, server *http.Server, finalSaver usecase.SnapshotSaver) error {
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

		if finalSaver != nil {
			if err := finalSaver.Execute(shutdownCtx); err != nil {
				return fmt.Errorf("final snapshot save: %w", err)
			}
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server close: %w", err)
		}
	}

	return nil
}

func buildLoggers() (*appLoggerRuntime, error) {
	baseZap, err := zap.NewProduction()

	if err != nil {
		return nil, err
	}

	httpLogger := baseZap.With(zap.String("component", "http"))
	appLogger := platformlogger.NewZapLogger(baseZap.With(zap.String("component", "application")))

	return &appLoggerRuntime{
		httpLogger: httpLogger,
		appLogger:  appLogger,
		cleanup:    func() { _ = baseZap.Sync() },
	}, nil
}

func buildStorageRuntime(ctx context.Context, cfg *cli.ServerConfig) (*storageRuntime, error) {
	switch cfg.StorageType {
	case cli.StorageTypeMemory:
		return buildMemoryStorageRuntime(), nil
	case cli.StorageTypeFile:
		return buildFileStorageRuntime(ctx, cfg)
	case cli.StorageTypePostgres:
		return buildPostgresStorageRuntime(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}

func buildPostgresStorageRuntime(ctx context.Context, cfg *cli.ServerConfig) (*storageRuntime, error) {
	if cfg.Postgres == nil {
		return nil, fmt.Errorf("postgres config is required")
	}

	postgresPool, err := platformpostgres.NewPool(ctx, cfg.Postgres)

	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := platformpostgres.Migrate(cfg.Postgres.ConnectionString(), "migrations"); err != nil {
		postgresPool.Close()
		return nil, fmt.Errorf("migrate postgres: %w", err)
	}

	metricRepo := storagepostgres.NewPostgresStorage(postgresPool)
	queryRepo := storagepostgres.NewQueryPostgresStorage(postgresPool)
	postgresChecker := platformpostgres.NewHealthChecker(postgresPool)
	batchRepo := storagepostgres.NewBatchRepository(postgresPool)

	return &storageRuntime{
		metricRepo:    metricRepo,
		queryRepo:     queryRepo,
		healthService: health.NewService(postgresChecker),
		snapshotSaver: nil,
		periodicSaver: nil,
		batchRepo:     batchRepo,
		cleanup:       postgresPool.Close,
	}, nil
}

func buildFileStorageRuntime(ctx context.Context, cfg *cli.ServerConfig) (*storageRuntime, error) {
	storage := memory.NewMemStorage()

	snapshotStore, err := snapshotfile.NewSnapshotStore(cfg.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("create snapshot store: %w", err)
	}

	snapshotMapper := mapper.NewMetricSnapshotMapper()

	saveSnapshotUC := usecase.NewSaveMetricSnapshotUseCase(
		storage,
		snapshotStore,
		snapshotMapper,
	)

	restoreUC := usecase.NewRestoreMetricUseCase(
		storage,
		snapshotStore,
		snapshotMapper,
	)

	if cfg.Restore {
		if err := restoreUC.Execute(ctx); err != nil {
			return nil, fmt.Errorf("restore metrics: %w", err)
		}
	}

	var snapshotSaver usecase.SnapshotSaver
	var periodicSaver usecase.SnapshotSaver

	if cfg.StoreInterval == 0 {
		snapshotSaver = saveSnapshotUC
	} else {
		periodicSaver = saveSnapshotUC
	}

	return &storageRuntime{
		metricRepo:    storage,
		queryRepo:     storage,
		healthService: health.NewService(),
		snapshotSaver: snapshotSaver,
		periodicSaver: periodicSaver,
		batchRepo:     storage,
		cleanup:       func() {},
	}, nil
}

func buildMemoryStorageRuntime() *storageRuntime {
	storage := memory.NewMemStorage()

	return &storageRuntime{
		metricRepo:    storage,
		queryRepo:     storage,
		healthService: health.NewService(),
		snapshotSaver: nil,
		periodicSaver: nil,
		batchRepo:     storage,
		cleanup:       func() {},
	}
}
