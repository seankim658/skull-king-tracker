package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	l "github.com/seankim658/skullking/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(accessLog zerolog.Logger, appLog zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

      requestID := uuid.NewString()
      requestSpecificAppLogger := appLog.With().Str(l.RequestIDKey, requestID).Logger()

      ctxWithLogger := l.NewContextWithLogger(r.Context(), requestSpecificAppLogger)
      r = r.WithContext(ctxWithLogger)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			accessLog.Info().
        Str(l.RequestIDKey, requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", r.URL.RawQuery).
				Int("status", rw.status).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Dur("duration_ms", duration).
				Msg("HTTP request processed")
		})
	}
}
