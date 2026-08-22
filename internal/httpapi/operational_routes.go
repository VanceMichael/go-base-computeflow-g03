package httpapi

import (
	"context"
	"encoding/json"
	"github.com/VanceMichael/harborflow/internal/audit"
	"github.com/VanceMichael/harborflow/internal/capacity"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/identity"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"net/http"
	"strconv"
	"time"
)

type OperationalRoutes struct {
	Store    *sqlite.Store
	Identity *identity.Service
	Audit    *audit.Service
	Capacity *capacity.Service
}

func NewOperationalRoutes(s *sqlite.Store, i *identity.Service, a *audit.Service, c *capacity.Service) *OperationalRoutes {
	return &OperationalRoutes{Store: s, Identity: i, Audit: a, Capacity: c}
}
func (o *OperationalRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/metrics", o.metrics)
	mux.HandleFunc("GET /api/v1/audit", o.audit)
	mux.HandleFunc("GET /api/v1/capacity/window", o.window)
}
func (o *OperationalRoutes) user(r *http.Request) (domain.User, error) {
	token := bearer(r)
	if token == "" {
		return domain.User{}, domain.ErrUnauthorized
	}
	return o.Identity.Authenticate(r.Context(), token, time.Now().UTC())
}
func (o *OperationalRoutes) metrics(w http.ResponseWriter, r *http.Request) {
	u, err := o.user(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := identity.Require(u.Role, identity.PermissionOperate); err != nil {
		writeError(w, err)
		return
	}
	m, err := o.Store.PortMetrics(r.Context(), u.PortID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
func (o *OperationalRoutes) audit(w http.ResponseWriter, r *http.Request) {
	u, err := o.user(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := identity.Require(u.Role, identity.PermissionAudit); err != nil {
		writeError(w, err)
		return
	}
	from, to, err := parseWindow(r)
	if err != nil {
		writeError(w, err)
		return
	}
	events, err := o.Audit.Timeline(r.Context(), u.PortID, from, to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "total": len(events)})
}
func (o *OperationalRoutes) window(w http.ResponseWriter, r *http.Request) {
	if _, err := o.user(r); err != nil {
		writeError(w, err)
		return
	}
	day := time.Now().UTC()
	if value := r.URL.Query().Get("day"); value != "" {
		parsed, err := time.Parse(time.DateOnly, value)
		if err != nil {
			writeError(w, domain.ErrInvalid)
			return
		}
		day = parsed
	}
	from, to := o.Capacity.Window(day)
	writeJSON(w, http.StatusOK, map[string]string{"from": from.Format(time.RFC3339), "to": to.Format(time.RFC3339)})
}
func parseWindow(r *http.Request) (time.Time, time.Time, error) {
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()
	if v := r.URL.Query().Get("from"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, time.Time{}, domain.ErrInvalid
		}
		from = parsed
	}
	if v := r.URL.Query().Get("to"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, time.Time{}, domain.ErrInvalid
		}
		to = parsed
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, domain.ErrInvalid
	}
	return from, to, nil
}
func pageFromHeaders(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.Header.Get("X-Page-Limit"))
	offset, _ := strconv.Atoi(r.Header.Get("X-Page-Offset"))
	if limit < 1 {
		limit = 50
	}
	return limit, offset
}
func decodeMap(r *http.Request) (map[string]string, error) {
	var m map[string]string
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return nil, domain.ErrInvalid
	}
	return m, nil
}
func callWithContext(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
