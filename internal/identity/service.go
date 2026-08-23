package identity

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/storage/sqlite"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Service struct {
	Store *sqlite.Store
	TTL   time.Duration
}

func New(s *sqlite.Store, ttl time.Duration) *Service { return &Service{Store: s, TTL: ttl} }
func hashToken(v string) string                       { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func (s *Service) Login(ctx context.Context, portID, email string, now time.Time) (domain.User, string, error) {
	u, err := s.Store.FindUserByEmail(ctx, portID, strings.ToLower(email))
	if err != nil {
		return u, "", domain.ErrUnauthorized
	}
	if !u.Active {
		return u, "", domain.ErrUnauthorized
	}
	raw := uuid.NewString() + uuid.NewString()
	x := domain.Session{ID: uuid.NewString(), UserID: u.ID, TokenHash: hashToken(raw), ExpiresAt: now.Add(s.TTL), CreatedAt: now}
	if err := s.Store.InsertSession(ctx, x); err != nil {
		return u, "", err
	}
	return u, raw, nil
}
func (s *Service) Authenticate(ctx context.Context, raw string, now time.Time) (domain.User, error) {
	u, err := s.Store.FindSessionUser(ctx, hashToken(raw), now)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return u, err
		}
		return u, domain.ErrUnauthorized
	}
	if !u.Active {
		return u, domain.ErrUnauthorized
	}
	return u, nil
}
func (s *Service) Logout(ctx context.Context, raw string, now time.Time) error {
	return s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return s.Store.RevokeSessionHash(ctx, tx, hashToken(raw), now)
	})
}
func (s *Service) Deactivate(ctx context.Context, actor, targetID string, now time.Time) error {
	if actor == "" || targetID == "" {
		return domain.ErrInvalid
	}
	return s.Store.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.Store.DeactivateUser(ctx, tx, targetID); err != nil {
			return err
		}
		if err := s.Store.RevokeSessions(ctx, tx, targetID, now); err != nil {
			return err
		}
		return nil
	})
}

var _ = errors.Is
