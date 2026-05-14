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

type QueryPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewQueryPostgresStorage(pool *pgxpool.Pool) *QueryPostgresStorage {
	return &QueryPostgresStorage{pool: pool}
}

func (q QueryPostgresStorage) ListGauges(ctx context.Context) ([]repository.GaugeSnapshot, error) {
	query := `
		SELECT name, gauge_value
		FROM metric
		WHERE type = $1
		ORDER BY name
	`

	var out []repository.GaugeSnapshot

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		rows, err := q.pool.Query(ctx, query, metricTypeGauge)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]repository.GaugeSnapshot, 0)

		for rows.Next() {
			var item repository.GaugeSnapshot
			if err := rows.Scan(&item.Name, &item.Value); err != nil {
				return err
			}

			result = append(result, item)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		out = result
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("list gauges: %w", err)
	}

	return out, nil
}

func (q QueryPostgresStorage) ListCounters(ctx context.Context) ([]repository.CounterSnapshot, error) {
	query := `
		SELECT name, counter_value
		FROM metric
		WHERE type = $1
		ORDER BY name
	`

	var out []repository.CounterSnapshot

	err := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		rows, err := q.pool.Query(ctx, query, metricTypeCounter)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]repository.CounterSnapshot, 0)

		for rows.Next() {
			var item repository.CounterSnapshot
			if err := rows.Scan(&item.Name, &item.Delta); err != nil {
				return err
			}

			result = append(result, item)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		out = result
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("list counters: %w", err)
	}

	return out, nil
}

func (q QueryPostgresStorage) FindGaugeByName(ctx context.Context, name metric.Name) (value float64, found bool, err error) {
	var id string
	var metricName string
	var valueRaw float64

	query := `
		SELECT id, name, gauge_value
		FROM metric
		WHERE metric.name = $1 AND metric.type = $2
	`

	errSql := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		return q.pool.QueryRow(
			ctx,
			query,
			name.String(),
			metricTypeGauge,
		).Scan(&id, &metricName, &valueRaw)
	})

	if errors.Is(errSql, pgx.ErrNoRows) {
		return 0, false, nil
	}

	if errSql != nil {
		return 0, false, fmt.Errorf("get gauge by name: %w", errSql)
	}

	gauge, err := metric.RestoreGauge(id, metricName, valueRaw)
	if err != nil {
		return 0, false, fmt.Errorf("restore gauge: %w", err)
	}

	return gauge.Value(), true, nil
}

func (q QueryPostgresStorage) FindCounterByName(ctx context.Context, name metric.Name) (delta int64, found bool, err error) {
	var id string
	var metricName string
	var value int64

	query := `
		SELECT id, name, counter_value
		FROM metric
		WHERE metric.name = $1 AND metric.type = $2
	`

	errSql := retry.Do(ctx, platformpostgres.IsRetriable, func() error {
		return q.pool.QueryRow(
			ctx,
			query,
			name.String(),
			metricTypeCounter,
		).Scan(&id, &metricName, &value)
	})

	if errors.Is(errSql, pgx.ErrNoRows) {
		return 0, false, nil
	}

	if errSql != nil {
		return 0, false, fmt.Errorf("get counter by name: %w", errSql)
	}

	counter, err := metric.RestoreCounter(id, metricName, value)
	if err != nil {
		return 0, false, fmt.Errorf("restore counter: %w", err)
	}

	return counter.Delta(), true, nil
}
