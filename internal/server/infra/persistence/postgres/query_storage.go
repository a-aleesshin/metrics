package postgres

import (
	"context"
	"errors"
	"fmt"

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

func (q QueryPostgresStorage) ListGauges() ([]repository.GaugeSnapshot, error) {
	query := `
		SELECT name, gauge_value
		FROM metric
		WHERE type = $1
		ORDER BY name
	`

	rows, err := q.pool.Query(context.Background(), query, metricTypeGauge)

	if err != nil {
		return nil, fmt.Errorf("list gauges: %w", err)
	}

	defer rows.Close()

	var out []repository.GaugeSnapshot

	for rows.Next() {
		var item repository.GaugeSnapshot
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, fmt.Errorf("scan gauge: %w", err)
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gauges: %w", err)
	}

	return out, nil
}

func (q QueryPostgresStorage) ListCounters() ([]repository.CounterSnapshot, error) {
	query := `
		SELECT name, counter_value
		FROM metric
		WHERE type = $1
		ORDER BY name
	`

	rows, err := q.pool.Query(context.Background(), query, metricTypeCounter)

	if err != nil {
		return nil, fmt.Errorf("list counters: %w", err)
	}

	defer rows.Close()

	var out []repository.CounterSnapshot

	for rows.Next() {
		var item repository.CounterSnapshot
		if err := rows.Scan(&item.Name, &item.Delta); err != nil {
			return nil, fmt.Errorf("scan counter: %w", err)
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate counters: %w", err)
	}

	return out, nil
}

func (q QueryPostgresStorage) FindGaugeByName(name metric.Name) (value float64, found bool, err error) {
	var id string
	var metricName string
	var valueRaw float64

	errSql := q.pool.QueryRow(
		context.Background(),
		"SELECT id, name, gauge_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
		name.String(),
		metricTypeGauge,
	).Scan(&id, &metricName, &valueRaw)

	if errors.Is(errSql, pgx.ErrNoRows) {
		return 0, false, nil
	}

	if errSql != nil {
		return 0, false, fmt.Errorf("get gauge by name: %w", err)
	}

	gauge, err := metric.RestoreGauge(id, metricName, valueRaw)

	if err != nil {
		return 0, false, fmt.Errorf("restore gauge: %w", err)
	}

	return gauge.Value(), true, nil
}

func (q QueryPostgresStorage) FindCounterByName(name metric.Name) (delta int64, found bool, err error) {
	var id string
	var metricName string
	var value int64

	errSql := q.pool.QueryRow(
		context.Background(),
		"SELECT id, name, gauge_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
		name.String(),
		metricTypeCounter,
	).Scan(&id, &metricName, &value)

	if errors.Is(errSql, pgx.ErrNoRows) {
		return 0, false, nil
	}

	if errSql != nil {
		return 0, false, fmt.Errorf("get gauge by name: %w", err)
	}

	counter, err := metric.RestoreCounter(id, metricName, value)

	if err != nil {
		return 0, false, fmt.Errorf("restore counter: %w", err)
	}

	return counter.Delta(), true, nil
}
