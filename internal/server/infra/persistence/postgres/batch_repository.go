package postgres

import (
	"context"
	"fmt"

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
	tx, err := b.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	gaugeSql := `
		INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (name, type)
		DO UPDATE SET
			gauge_value = EXCLUDED.gauge_value,
			counter_value = NULL,
			updated_at = now()
	`

	for _, gauge := range batch.Gauges {
		_, err := tx.Exec(
			ctx,
			gaugeSql,
			gauge.Id().String(),
			gauge.Name().String(),
			metricTypeGauge,
			gauge.Value(),
		)

		if err != nil {
			return fmt.Errorf("save gauge %q in batch: %w", gauge.Name().String(), err)
		}
	}

	counterSql := `
	INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, NULL, $4)
		ON CONFLICT (name, type)
		DO UPDATE SET
			counter_value = metric.counter_value + EXCLUDED.counter_value,
			gauge_value = NULL,
			updated_at = now()
	`

	for _, counter := range batch.Counters {
		_, err := tx.Exec(
			ctx,
			counterSql,
			counter.Id().String(),
			counter.Name().String(),
			metricTypeCounter,
			counter.Delta(),
		)

		if err != nil {
			return fmt.Errorf("save counter %q in batch: %w", counter.Name().String(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit batch tx: %w", err)
	}

	return nil
}
