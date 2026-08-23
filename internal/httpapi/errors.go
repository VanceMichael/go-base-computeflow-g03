package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal"
	message := "internal server error"
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
		message = "authentication required"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
		message = "resource not found"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		code = "conflict"
		message = "operation conflicts with current state"
	case errors.Is(err, domain.ErrInvalid):
		status = http.StatusUnprocessableEntity
		code = "invalid"
		message = "operation is not valid"
	case errors.Is(err, domain.ErrUnavailable):
		status = http.StatusServiceUnavailable
		code = "unavailable"
		message = "dependency is temporarily unavailable"
	}
	writeJSON(w, status, errorBody{Code: code, Message: message})
}
