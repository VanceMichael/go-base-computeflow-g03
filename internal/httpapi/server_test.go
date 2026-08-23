package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/VanceMichael/computeflow/internal/flow"
	"github.com/VanceMichael/computeflow/internal/httpapi"
	"github.com/VanceMichael/computeflow/internal/identity"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthEndpointReturnsJSON(t *testing.T) {
	f := testsupport.New(t)
	h := httpapi.New(f.Store, identity.New(f.Store, time.Hour), flow.New(f.Store)).Handler()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
func TestReadyEndpointChecksDatabase(t *testing.T) {
	f := testsupport.New(t)
	h := httpapi.New(f.Store, identity.New(f.Store, time.Hour), flow.New(f.Store)).Handler()
	r := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}
func TestMeRequiresAuthentication(t *testing.T) {
	f := testsupport.New(t)
	h := httpapi.New(f.Store, identity.New(f.Store, time.Hour), flow.New(f.Store)).Handler()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
func TestLoginAndMeRoundTrip(t *testing.T) {
	f := testsupport.New(t)
	id := identity.New(f.Store, time.Hour)
	h := httpapi.New(f.Store, id, flow.New(f.Store)).Handler()
	payload, _ := json.Marshal(map[string]string{"port_id": f.Port.ID, "email": f.User.Email})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Token == "" {
		t.Fatalf("%+v %v", result, err)
	}
	r = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r.Header.Set("Authorization", "Bearer "+result.Token)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("me %d %s", w.Code, w.Body.String())
	}
}
func TestCreateRunRequiresCoordinatorRole(t *testing.T) {
	f := testsupport.New(t)
	id := identity.New(f.Store, time.Hour)
	h := httpapi.New(f.Store, id, flow.New(f.Store)).Handler()
	_, token, err := id.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]string{"port_id": f.Port.ID, "name": "test"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/runs", bytes.NewReader(payload))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
func TestLogoutReturnsNoContent(t *testing.T) {
	f := testsupport.New(t)
	id := identity.New(f.Store, time.Hour)
	h := httpapi.New(f.Store, id, flow.New(f.Store)).Handler()
	_, token, err := id.Login(context.Background(), f.Port.ID, f.User.Email, f.Now)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("%d", w.Code)
	}
}
