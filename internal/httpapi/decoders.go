package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
	"io"
	"net/http"
	"strings"
)

func decodeObject(r *http.Request, dst any) error {
	if r.Body == nil {
		return fmt.Errorf("%w: empty body", domain.ErrInvalid)
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid json", domain.ErrInvalid)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("%w: multiple json values", domain.ErrInvalid)
	}
	return nil
}
func requireHeader(r *http.Request, name string) (string, error) {
	v := strings.TrimSpace(r.Header.Get(name))
	if v == "" {
		return "", fmt.Errorf("%w: missing %s", domain.ErrInvalid, name)
	}
	return v, nil
}
func acceptJSON(r *http.Request) bool {
	v := r.Header.Get("Accept")
	return v == "" || strings.Contains(v, "application/json")
}
func writeCreated(w http.ResponseWriter, value any) { writeJSON(w, http.StatusCreated, value) }
