package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
)

func Migrate(dsn string, migrationsPath string) error {
	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
