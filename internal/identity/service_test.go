package identity_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/identity"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"testing"
	"time"
)

func TestLoginAuthenticatesActiveUser(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Hour)
	u, token, err := s.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil || u.ID != f.User.ID || token == "" {
		t.Fatalf("%+v %q %v", u, token, err)
	}
	got, err := s.Authenticate(context.Background(), token, f.Now.Add(time.Minute))
	if err != nil || got.ID != f.User.ID {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestLoginRejectsWrongPort(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Hour)
	if _, _, err := s.Login(context.Background(), "foreign", f.User.Email, f.Now); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestLoginRejectsInactiveUser(t *testing.T) {
	f := testsupport.New(t)
	if _, err := f.Store.DB.Exec(`UPDATE users SET active=0 WHERE id=?`, f.User.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := identity.New(f.Store, time.Hour).Login(context.Background(), f.Port.ID, f.User.Email, f.Now); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestLogoutRevokesSession(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Hour)
	_, token, err := s.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Logout(context.Background(), token, f.Now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(context.Background(), token, f.Now); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestExpiredSessionCannotAuthenticate(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Minute)
	_, token, err := s.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(context.Background(), token, f.Now.Add(2*time.Minute)); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestDeactivateUserRevokesExistingSessions(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Hour)
	_, token, err := s.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Deactivate(context.Background(), "admin", f.User.ID, f.Now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(context.Background(), token, f.Now); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestSessionNotExpiredAtExactBoundary(t *testing.T) {
	f := testsupport.New(t)
	s := identity.New(f.Store, time.Hour)
	_, token, err := s.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(context.Background(), token, f.Now.Add(time.Hour)); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("boundary should be expired: %v", err)
	}
}
