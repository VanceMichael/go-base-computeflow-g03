package httpapi

import (
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"net/http"
)

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalid):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
func publicMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return "authentication required"
	case errors.Is(err, domain.ErrNotFound):
		return "resource not found"
	case errors.Is(err, domain.ErrConflict):
		return "operation conflicts with current state"
	case errors.Is(err, domain.ErrInvalid):
		return "request is not valid"
	case errors.Is(err, domain.ErrUnavailable):
		return "dependency is temporarily unavailable"
	default:
		return "internal server error"
	}
}
