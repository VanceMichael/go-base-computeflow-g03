package httpapi

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"net/http"
	"strings"
)

func validateMethod(r *http.Request, want string) error {
	if r.Method != want {
		return fmt.Errorf("%w: method %s", domain.ErrInvalid, r.Method)
	}
	return nil
}
func validatePortHeader(r *http.Request, expected string) error {
	if expected == "" {
		return nil
	}
	if strings.TrimSpace(r.Header.Get("X-Port-ID")) != expected {
		return fmt.Errorf("%w: port scope", domain.ErrUnauthorized)
	}
	return nil
}
func requireContentType(r *http.Request) error {
	if r.Body == nil {
		return fmt.Errorf("%w: body", domain.ErrInvalid)
	}
	ct := r.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("%w: content type", domain.ErrInvalid)
	}
	return nil
}
