package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

var ErrMalformedToken = errors.New("malformed token")

type Signer struct{ secret []byte }

func NewSigner(secret string) *Signer { return &Signer{secret: []byte(secret)} }
func (s *Signer) Sign(subject string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(subject))
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(subject)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}
func (s *Signer) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", ErrMalformedToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrMalformedToken
	}
	want, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrMalformedToken
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(raw)
	if !hmac.Equal(mac.Sum(nil), want) {
		return "", ErrMalformedToken
	}
	return string(raw), nil
}
