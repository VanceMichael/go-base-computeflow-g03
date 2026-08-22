package httpapi

import (
	"encoding/json"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/flow"
	"github.com/VanceMichael/harborflow/internal/identity"
	"github.com/VanceMichael/harborflow/internal/middleware"
	"github.com/VanceMichael/harborflow/internal/storage/sqlite"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	Store    *sqlite.Store
	Identity *identity.Service
	Flow     *flow.Service
}

func New(s *sqlite.Store, i *identity.Service, f *flow.Service) *Server {
	return &Server{Store: s, Identity: i, Flow: f}
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("POST /api/v1/login", s.login)
	mux.HandleFunc("POST /api/v1/logout", s.logout)
	mux.HandleFunc("GET /api/v1/me", s.me)
	mux.HandleFunc("POST /api/v1/runs", s.createRun)
	return mux
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.Ping(r.Context()); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type loginRequest struct {
	PortID string `json:"port_id"`
	Email  string `json:"email"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, domain.ErrInvalid)
		return
	}
	u, token, err := s.Identity.Login(r.Context(), in.PortID, in.Email, time.Now().UTC())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	token := bearer(r)
	if token == "" {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	if err := s.Identity.Logout(r.Context(), token, time.Now().UTC()); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.auth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type createRunRequest struct {
	PortID string `json:"port_id"`
	Name   string `json:"name"`
}

func (s *Server) createRun(w http.ResponseWriter, r *http.Request) {
	u, err := s.auth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if u.Role != domain.RoleCoordinator {
		writeError(w, domain.ErrUnauthorized)
		return
	}
	var in createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, domain.ErrInvalid)
		return
	}
	run, err := s.Flow.CreateRun(r.Context(), in.PortID, in.Name, u.ID, time.Now().UTC())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}
func (s *Server) auth(r *http.Request) (domain.User, error) {
	token := bearer(r)
	if token == "" {
		return domain.User{}, domain.ErrUnauthorized
	}
	return s.Identity.Authenticate(r.Context(), token, time.Now().UTC())
}
func bearer(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if !strings.HasPrefix(v, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
}

var _ = middleware.RequestID
