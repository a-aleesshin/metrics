package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgerrcode.IsConnectionException(pgErr.Code)
}
