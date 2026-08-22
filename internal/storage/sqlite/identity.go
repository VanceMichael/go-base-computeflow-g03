package sqlite

import (
	"context"
	"database/sql"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

func (s *Store) InsertUser(ctx context.Context, u domain.User) error {
	return exec(s.DB, ctx, `INSERT INTO users(id,port_id,email,display_name,role,active,created_at) VALUES(?,?,?,?,?,?,?)`, u.ID, u.PortID, u.Email, u.DisplayName, string(u.Role), boolInt(u.Active), stamp(u.CreatedAt))
}
func (s *Store) FindUserByEmail(ctx context.Context, portID, email string) (domain.User, error) {
	var u domain.User
	var active int
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT id,port_id,email,display_name,role,active,created_at FROM users WHERE port_id=? AND email=?`, portID, email).Scan(&u.ID, &u.PortID, &u.Email, &u.DisplayName, &u.Role, &active, &created)
	if err != nil {
		return u, err
	}
	u.Active = active == 1
	u.CreatedAt, _ = scanTime(created)
	return u, nil
}
func (s *Store) InsertSession(ctx context.Context, x domain.Session) error {
	var revoked any
	if x.RevokedAt != nil {
		revoked = stamp(*x.RevokedAt)
	}
	return exec(s.DB, ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at,revoked_at,created_at) VALUES(?,?,?,?,?,?)`, x.ID, x.UserID, x.TokenHash, stamp(x.ExpiresAt), revoked, stamp(x.CreatedAt))
}
func (s *Store) FindSessionUser(ctx context.Context, tokenHash string, now time.Time) (domain.User, error) {
	var u domain.User
	var active int
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT u.id,u.port_id,u.email,u.display_name,u.role,u.active,u.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND s.revoked_at IS NULL`, tokenHash, stamp(now)).Scan(&u.ID, &u.PortID, &u.Email, &u.DisplayName, &u.Role, &active, &created)
	if err != nil {
		return u, err
	}
	u.Active = active == 1
	u.CreatedAt, _ = scanTime(created)
	return u, nil
}
func (s *Store) RevokeSessions(ctx context.Context, tx *sql.Tx, userID string, now time.Time) error {
	return exec(tx, ctx, `UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`, stamp(now), userID)
}
func (s *Store) RevokeSessionHash(ctx context.Context, tx *sql.Tx, tokenHash string, now time.Time) error {
	return exec(tx, ctx, `UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL`, stamp(now), tokenHash)
}
func (s *Store) DeactivateUser(ctx context.Context, tx *sql.Tx, userID string) error {
	return exec(tx, ctx, `UPDATE users SET active=0 WHERE id=? AND active=1`, userID)
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
