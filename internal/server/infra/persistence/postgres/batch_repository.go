package postgres

import (
	"context"
	"fmt"

	platformpostgres "github.com/a-aleesshin/metrics/internal/platform/db/postgres"
	"github.com/a-aleesshin/metrics/internal/platform/retry"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BatchRepository struct {
	pool *pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool: pool}
}

func (b BatchRepository) UpdateBatch(ctx context.Context, batch repository.MetricBatch) error {
	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		if err := b.updateBatch(ctx, batch); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("update metric batch: %w", err)
	}

	return nil
}

func (b BatchRepository) updateBatch(ctx context.Context, batch repository.MetricBatch) error {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin batch tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const gaugeSQL = `
		INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (name, type)
		DO UPDATE SET
			gauge_value = EXCLUDED.gauge_value,
			counter_value = NULL,
			updated_at = now()
	`

	for _, gauge := range batch.Gauges {
		if _, err := tx.Exec(
			ctx,
			gaugeSQL,
			gauge.Id().String(),
			gauge.Name().String(),
			metricTypeGauge,
			gauge.Value(),
		); err != nil {
			return fmt.Errorf("save gauge %q in batch: %w", gauge.Name().String(), err)
		}
	}

	const counterSQL = `
		INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, NULL, $4)
		ON CONFLICT (name, type)
		DO UPDATE SET
			counter_value = metric.counter_value + EXCLUDED.counter_value,
			gauge_value = NULL,
			updated_at = now()
	`

	for _, counter := range batch.Counters {
		if _, err := tx.Exec(
			ctx,
			counterSQL,
			counter.Id().String(),
			counter.Name().String(),
			metricTypeCounter,
			counter.Delta(),
		); err != nil {
			return fmt.Errorf("save counter %q in batch: %w", counter.Name().String(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit batch tx: %w", err)
	}

	return nil
}
