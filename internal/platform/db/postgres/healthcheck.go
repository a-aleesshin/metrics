package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthChecker struct {
	pool *pgxpool.Pool
}

func NewHealthChecker(p *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{pool: p}
}

func (checker *HealthChecker) Name() string {
	return "postgres"
}

func (checker *HealthChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.pool.Ping(ctx)

	if err != nil {
		checker.pool.Close()
		return err
	}

	return nil
}
