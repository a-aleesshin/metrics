package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	dbPool, err := pgxpool.New(context.Background(), cfg.ConnectionString())

	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}

	return dbPool, nil
}
