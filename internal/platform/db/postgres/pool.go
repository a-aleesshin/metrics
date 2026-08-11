// Package postgres содержит инфраструктуру PostgreSQL: конфигурацию DSN,
// пул соединений pgx, миграции, healthcheck и классификацию retriable-ошибок.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул соединений pgxpool по конфигурации; cfg не может быть nil.
func NewPool(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	dbPool, err := pgxpool.New(ctx, cfg.ConnectionString())

	if err != nil {
		return nil, err
	}

	return dbPool, nil
}
