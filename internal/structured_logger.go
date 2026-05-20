package internal

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func StructuredLoggerFromContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := LoggerFromContext(r.Context())
			recordRequest(logger, next, w, r)
		})
	}
}

// StructuredLogger returns a middleware that logs each request in structured (JSON) form using slog.
func StructuredLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recordRequest(logger, next, w, r)
		})
	}
}

func recordRequest(logger *slog.Logger, next http.Handler, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Use a wrapper to get the status code
	ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
	next.ServeHTTP(ww, r)

	logHealth := os.Getenv("LOG_HEALTH") == "true"

	if r.URL.Path == "/health" && !logHealth {
		return
	}
	duration := time.Since(start)

	logger.Info("WEB_REQUEST",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", ww.status),
		slog.String("remote", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("duration_ms", duration.Milliseconds()),
	)
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
