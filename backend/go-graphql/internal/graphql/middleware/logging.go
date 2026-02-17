package middleware

import (
	"net/http"
	"time"

	platformlogger "github.com/faizp/zenlist/backend/go-graphql/internal/platform/logger"
)

func Logging(logger *platformlogger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", RequestIDFromContext(r.Context()),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
