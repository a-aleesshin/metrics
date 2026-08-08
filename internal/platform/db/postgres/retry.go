package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// IsRetriable сообщает, стоит ли повторять операцию после ошибки PostgreSQL:
// сбои сериализации, дедлоки, обрывы соединения и перегрузка по подключениям.
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.AdminShutdown,
		pgerrcode.TooManyConnections:
		return true
	default:
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
}
