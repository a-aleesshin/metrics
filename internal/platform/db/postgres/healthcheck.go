package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthChecker проверяет доступность PostgreSQL через Ping пула.
type HealthChecker struct {
	pool *pgxpool.Pool
}

// NewHealthChecker создаёт проверку здоровья для пула соединений.
func NewHealthChecker(p *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{pool: p}
}

// Name возвращает имя проверки — "postgres".
func (checker *HealthChecker) Name() string {
	return "postgres"
}

// Check пингует базу с таймаутом 2 секунды.
func (checker *HealthChecker) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return checker.pool.Ping(ctx)
}
