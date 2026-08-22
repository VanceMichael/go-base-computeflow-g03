package security_test

import (
	"github.com/VanceMichael/harborflow/internal/security"
	"testing"
)

func TestSignerRoundTrip(t *testing.T) {
	s := security.NewSigner("test-secret")
	token := s.Sign("user-1")
	got, err := s.Verify(token)
	if err != nil || got != "user-1" {
		t.Fatalf("%q %v", got, err)
	}
}
func TestSignerRejectsTampering(t *testing.T) {
	s := security.NewSigner("test-secret")
	token := s.Sign("user-1") + "x"
	if _, err := s.Verify(token); err == nil {
		t.Fatal("tampered token accepted")
	}
}
func TestSignerRejectsMissingParts(t *testing.T) {
	if _, err := security.NewSigner("x").Verify("bad"); err == nil {
		t.Fatal("malformed token accepted")
	}
}
func TestSignerUsesSecretAsOwnershipBoundary(t *testing.T) {
	a := security.NewSigner("a")
	b := security.NewSigner("b")
	if _, err := b.Verify(a.Sign("user")); err == nil {
		t.Fatal("different secret accepted")
	}
}
