package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "ordinary error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "connection exception",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionException},
			want: true,
		},
		{
			name: "connection failure",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			want: true,
		},
		{
			name: "unique violation is not retriable",
			err:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetriable(tt.err)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
