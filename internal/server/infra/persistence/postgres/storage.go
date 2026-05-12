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

func (p PostgresStorage) GetGaugeByName(name metric.Name) (*metric.Gauge, error) {
	var id string
	var metricName string
	var value float64

	err := p.pool.QueryRow(
		context.Background(),
		"SELECT id, name, gauge_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
		name.String(),
		metricTypeGauge,
	).Scan(&id, &metricName, &value)

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

func (p PostgresStorage) GetCounterByName(name metric.Name) (*metric.Counter, error) {
	var id string
	var metricName string
	var value int64

	err := p.pool.QueryRow(
		context.Background(),
		"SELECT id, name, counter_value FROM metric WHERE metric.name = $1 AND metric.type = $2",
		name.String(),
		metricTypeCounter,
	).Scan(&id, &metricName, &value)

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

func (p PostgresStorage) SaveGauge(gauge *metric.Gauge) error {
	sql := `
			INSERT INTO metric (id, name, type, gauge_value, counter_value) VALUES ($1, $2, $3, $4, NULL)
			ON CONFLICT (name, type)
			DO UPDATE SET
				gauge_value = EXCLUDED.gauge_value,
				counter_value = NULL,
				updated_at = now()
			`

	_, err := p.pool.Exec(
		context.Background(),
		sql,
		gauge.Id().String(),
		gauge.Name().String(),
		metricTypeGauge,
		gauge.Value(),
	)

	if err != nil {
		return fmt.Errorf("save gauge: %w", err)
	}

	return nil
}

func (p PostgresStorage) SaveCounter(counter *metric.Counter) error {
	sql := `
			INSERT INTO metric (id, name, type, gauge_value, counter_value) VALUES ($1, $2, $3, NULL, $4)
			ON CONFLICT (name, type)
			DO UPDATE SET
				counter_value = EXCLUDED.counter_value,
				gauge_value = NULL,
				updated_at = now()
			`

	_, err := p.pool.Exec(
		context.Background(),
		sql,
		counter.Id().String(),
		counter.Name().String(),
		metricTypeCounter,
		counter.Delta(),
	)

	if err != nil {
		return fmt.Errorf("save gauge: %w", err)
	}

	return nil
}

func (p PostgresStorage) Save(metrics []repository.MetricSnapshot) error {
	ctx := context.Background()

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
			counter_value = EXCLUDED.counter_value,
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

func (p PostgresStorage) Load() ([]repository.MetricSnapshot, error) {
	query := `SELECT id, type, gauge_value, counter_value FROM metric ORDER BY type, name`

	rows, err := p.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("load metric snapshots: %w", err)
	}
	defer rows.Close()

	snapshots := make([]repository.MetricSnapshot, 0)

	for rows.Next() {
		var snapshot repository.MetricSnapshot

		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.Type,
			&snapshot.Value,
			&snapshot.Delta,
		); err != nil {
			return nil, fmt.Errorf("scan metric snapshot: %w", err)
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metric snapshots: %w", err)
	}

	return snapshots, nil
}
