package middleware

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type requestKey struct{}

func RequestID(ctx context.Context) string { v, _ := ctx.Value(requestKey{}).(string); return v }
func Chain(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestKey{}, id)
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", "request_id", id, "panic", recovered, "stack", string(debug.Stack()))
				http.Error(w, `{"code":"internal","message":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
		logger.Info("request complete", "request_id", id, "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}
