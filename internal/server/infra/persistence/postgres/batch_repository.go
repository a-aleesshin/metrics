package postgres

import (
	"context"
	"fmt"

	platformpostgres "github.com/a-aleesshin/metrics/internal/platform/db/postgres"
	"github.com/a-aleesshin/metrics/internal/platform/retry"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BatchRepository struct {
	pool *pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool: pool}
}

func (b BatchRepository) UpdateBatch(ctx context.Context, batch repository.MetricBatch) error {
	if len(batch.Gauges) == 0 && len(batch.Counters) == 0 {
		return nil
	}

	if err := b.withRetryTx(ctx, func(tx pgx.Tx) error {
		return b.updateBatch(ctx, tx, batch)
	}); err != nil {
		return fmt.Errorf("update metric batch: %w", err)
	}

	return nil
}

func (b BatchRepository) updateBatch(ctx context.Context, tx pgx.Tx, batch repository.MetricBatch) error {
	const gaugeSQL = `
		INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (name, type)
		DO UPDATE SET
			gauge_value = EXCLUDED.gauge_value,
			counter_value = NULL,
			updated_at = now()
	`

	pgxBatch := &pgx.Batch{}

	for _, gauge := range batch.Gauges {
		pgxBatch.Queue(
			gaugeSQL,
			gauge.Id().String(),
			gauge.Name().String(),
			metricTypeGauge,
			gauge.Value(),
		)
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
		pgxBatch.Queue(
			counterSQL,
			counter.Id().String(),
			counter.Name().String(),
			metricTypeCounter,
			counter.Delta(),
		)
	}

	results := tx.SendBatch(ctx, pgxBatch)

	for i, gauge := range batch.Gauges {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()
			return fmt.Errorf("save gauge %q in batch item %d: %w", gauge.Name().String(), i, err)
		}
	}

	for i, counter := range batch.Counters {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()
			return fmt.Errorf("save counter %q in batch item %d: %w", counter.Name().String(), i, err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close batch results: %w", err)
	}

	return nil
}

func (b BatchRepository) withRetryTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		tx, err := b.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin batch tx: %w", err)
		}

		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				fmt.Printf("rollback batch tx: %v\n", err)
			}
		}()

		if err := fn(tx); err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit batch tx: %w", err)
		}

		return nil
	})
}
