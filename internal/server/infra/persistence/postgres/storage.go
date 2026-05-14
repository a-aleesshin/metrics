package postgres

import (
	"context"
	"errors"
	"fmt"

	platformpostgres "github.com/a-aleesshin/metrics/internal/platform/db/postgres"
	"github.com/a-aleesshin/metrics/internal/platform/retry"
	"github.com/a-aleesshin/metrics/internal/server/application/port/repository"
	"github.com/a-aleesshin/metrics/internal/server/domain/metric"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

func (p PostgresStorage) GetGaugeByName(ctx context.Context, name metric.Name) (*metric.Gauge, error) {
	var id string
	var metricName string
	var value float64

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		return p.pool.QueryRow(
			ctx,
			"SELECT id, name, gauge_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
			name.String(),
			metricTypeGauge,
		).Scan(&id, &metricName, &value)
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get gauge by name: %w", err)
	}

	gauge, err := metric.RestoreGauge(id, metricName, value)

	if err != nil {
		return nil, fmt.Errorf("restore gauge: %w", err)
	}

	return gauge, nil
}

func (p PostgresStorage) GetCounterByName(ctx context.Context, name metric.Name) (*metric.Counter, error) {
	var id string
	var metricName string
	var value int64

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		return p.pool.QueryRow(
			ctx,
			"SELECT id, name, counter_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
			name.String(),
			metricTypeCounter,
		).Scan(&id, &metricName, &value)
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get gauge by name: %w", err)
	}

	counter, err := metric.RestoreCounter(id, metricName, value)

	if err != nil {
		return nil, fmt.Errorf("restore counter: %w", err)
	}

	return counter, nil
}

func (p PostgresStorage) SaveGauge(ctx context.Context, gauge *metric.Gauge) error {
	sql := `
			INSERT INTO metric (id, name, type, gauge_value, counter_value) VALUES ($1, $2, $3, $4, NULL)
			ON CONFLICT (name, type)
			DO UPDATE SET
				gauge_value = EXCLUDED.gauge_value,
				counter_value = NULL,
				updated_at = now()
			`

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		_, err := p.pool.Exec(
			ctx,
			sql,
			gauge.Id().String(),
			gauge.Name().String(),
			metricTypeGauge,
			gauge.Value(),
		)
		return err
	})

	if err != nil {
		return fmt.Errorf("save gauge: %w", err)
	}

	return nil
}

func (p PostgresStorage) SaveCounter(ctx context.Context, counter *metric.Counter) error {
	sql := `
			INSERT INTO metric (id, name, type, gauge_value, counter_value) VALUES ($1, $2, $3, NULL, $4)
			ON CONFLICT (name, type)
			DO UPDATE SET
				counter_value = EXCLUDED.counter_value,
				gauge_value = NULL,
				updated_at = now()
			`

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		_, err := p.pool.Exec(
			ctx,
			sql,
			counter.Id().String(),
			counter.Name().String(),
			metricTypeCounter,
			counter.Delta(),
		)
		return err
	})

	if err != nil {
		return fmt.Errorf("save counter: %w", err)
	}

	return nil
}

func (p PostgresStorage) Save(ctx context.Context, metrics []repository.MetricSnapshot) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin save snapshots tx: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			fmt.Printf("rollback save snapshots tx: %v\n", err)
		}
	}(tx, ctx)

	query := `
		INSERT INTO metric (id, name, type, gauge_value, counter_value)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name, type)
		DO UPDATE SET
			gauge_value = EXCLUDED.gauge_value,
			counter_value = metric.counter_value + EXCLUDED.counter_value,
			updated_at = now()
	`

	for _, snapshot := range metrics {
		if _, err := tx.Exec(
			ctx,
			query,
			snapshot.ID,
			snapshot.ID,
			snapshot.Type,
			snapshot.Value,
			snapshot.Delta,
		); err != nil {
			return fmt.Errorf("save metric snapshot %q: %w", snapshot.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit save snapshots tx: %w", err)
	}

	return nil
}

func (p PostgresStorage) Load(ctx context.Context) ([]repository.MetricSnapshot, error) {
	query := `SELECT id, type, gauge_value, counter_value FROM metric ORDER BY type, name`

	var snapshots []repository.MetricSnapshot

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		rows, err := p.pool.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]repository.MetricSnapshot, 0)

		for rows.Next() {
			var snapshot repository.MetricSnapshot
			if err := rows.Scan(&snapshot.ID, &snapshot.Type, &snapshot.Value, &snapshot.Delta); err != nil {
				return err
			}
			result = append(result, snapshot)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		snapshots = result
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("load metric snapshots: %w", err)
	}

	return snapshots, nil
}
