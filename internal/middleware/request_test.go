package middleware_test

import (
	"github.com/VanceMichael/harborflow/internal/middleware"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAddsRequestIDWhenMissing(t *testing.T) {
	h := middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.RequestID(r.Context()) == "" {
			t.Fatal("request id missing")
		}
		w.WriteHeader(http.StatusNoContent)
	}), slog.Default())
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatal(w.Code)
	}
}
func TestMiddlewarePreservesProvidedRequestID(t *testing.T) {
	h := middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.RequestID(r.Context()) != "provided" {
			t.Fatal("request id changed")
		}
	}), slog.Default())
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", "provided")
	h.ServeHTTP(httptest.NewRecorder(), r)
}
func TestMiddlewareRecoversPanic(t *testing.T) {
	h := middleware.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }), slog.Default())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}
