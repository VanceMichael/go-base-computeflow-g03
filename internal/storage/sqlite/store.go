package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB  *sql.DB
	now func() time.Time
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	s := &Store{DB: db, now: time.Now}
	if err := s.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func OpenMemory() (*Store, error) {
	db, err := sql.Open("sqlite", "file:harborflow-test-"+uuid.NewString()+"?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{DB: db, now: time.Now}
	if err := s.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error                  { return s.DB.Close() }
func (s *Store) Now() time.Time                { return s.now().UTC() }
func (s *Store) SetClock(now func() time.Time) { s.now = now }

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	var version int
	_ = s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&version)
	if version < 1 {
		body, err := readMigration()
		if err != nil {
			return err
		}
		if _, err = s.DB.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("migration 1: %w", err)
		}
		if _, err = s.DB.ExecContext(ctx, `INSERT INTO schema_migrations(version,applied_at) VALUES(1,?)`, s.Now().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	return nil
}

func readMigration() ([]byte, error) {
	candidates := []string{"migrations/001_init.sql", "../migrations/001_init.sql", "../../migrations/001_init.sql", "../../../migrations/001_init.sql", "../../../../migrations/001_init.sql", "/app/migrations/001_init.sql"}
	for _, name := range candidates {
		if body, err := os.ReadFile(filepath.Clean(name)); err == nil {
			return body, nil
		}
	}
	return nil, fmt.Errorf("migration file not found")
}

func (s *Store) WithTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func IsConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "constraint") || strings.Contains(strings.ToLower(err.Error()), "unique")
}
func Normalize(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainNotFound{}
	}
	return err
}

type domainNotFound struct{}

func (domainNotFound) Error() string { return "not found" }

func scanTime(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
func nullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	t, err := scanTime(value.String)
	if err != nil {
		return nil
	}
	return &t
}
