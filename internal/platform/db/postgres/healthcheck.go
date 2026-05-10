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
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return checker.pool.Ping(ctx)
}
