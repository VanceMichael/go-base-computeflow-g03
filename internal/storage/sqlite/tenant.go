package sqlite

import (
	"context"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/google/uuid"
)

func (s *Store) InsertPort(ctx context.Context, p domain.Port) error {
	return exec(s.DB, ctx, `INSERT INTO ports(id,code,name,timezone,created_at) VALUES(?,?,?,?,?)`, p.ID, p.Code, p.Name, p.Timezone, stamp(p.CreatedAt))
}
func (s *Store) GetPort(ctx context.Context, id string) (domain.Port, error) {
	var p domain.Port
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT id,code,name,timezone,created_at FROM ports WHERE id=?`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Timezone, &created)
	if err != nil {
		return p, err
	}
	p.CreatedAt, _ = scanTime(created)
	return p, nil
}
func (s *Store) EnsureDemoPort(ctx context.Context) (domain.Port, error) {
	p := domain.Port{ID: uuid.NewString(), Code: "HKG-SZX", Name: "ComputeFlow Joint Port", Timezone: "Asia/Shanghai", CreatedAt: s.Now()}
	if _, err := s.DB.ExecContext(ctx, `INSERT OR IGNORE INTO ports(id,code,name,timezone,created_at) VALUES(?,?,?,?,?)`, p.ID, p.Code, p.Name, p.Timezone, stamp(p.CreatedAt)); err != nil {
		return p, err
	}
	var created string
	if err := s.DB.QueryRowContext(ctx, `SELECT id,code,name,timezone,created_at FROM ports WHERE code=?`, `HKG-SZX`).Scan(&p.ID, &p.Code, &p.Name, &p.Timezone, &created); err != nil {
		return p, err
	}
	p.CreatedAt, _ = scanTime(created)
	return p, nil
}
