package sqlite

import (
	"context"
	"database/sql"
	"time"
)

type Executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func exec(ex Executor, ctx context.Context, query string, args ...any) error {
	_, err := ex.ExecContext(ctx, query, args...)
	return err
}
func rowsAffected(ex Executor, ctx context.Context, query string, args ...any) (int64, error) {
	r, err := ex.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
