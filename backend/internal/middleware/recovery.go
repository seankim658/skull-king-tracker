package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"

	l "github.com/seankim658/skullking/internal/logger"
)

// Recovers from panics in HTTP handlers
func RecoveryMiddleware(appLog zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					appLog.Error().
						Interface("panic_error", err).
						Bytes("stack_trace", debug.Stack()).
						Str("request_method", r.Method).
						Str("request_path", r.URL.Path).
						Msg("Recovered from panic in HTTP handler")

					http.Error(w, l.InternalServerError, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
