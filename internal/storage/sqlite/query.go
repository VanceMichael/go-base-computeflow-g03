package sqlite

import (
	"context"
	"database/sql"
)

func (s *Store) Ping(ctx context.Context) error { return s.DB.PingContext(ctx) }
func CloseRows(rows *sql.Rows) error {
	if rows == nil {
		return nil
	}
	return rows.Close()
}
