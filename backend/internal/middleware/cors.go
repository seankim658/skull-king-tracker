package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"

	cf "github.com/seankim658/skullking/internal/config"
)

func CorsMiddleware(cfg *cf.Config) func(http.Handler) http.Handler {
	allowedOrigins := []string{}
	if cfg.AppEnv == "production" {
		allowedOrigins = append(allowedOrigins, cfg.AppBaseURL)
	} else {
		allowedOrigins = append(allowedOrigins, "http://localhost:5173")
		allowedOrigins = append(allowedOrigins, "http://localhost:3000")
		allowedOrigins = append(allowedOrigins, "http://localhost:"+cfg.APIPort)
	}

	allowedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodPatch,
	}

	allowedHeaders := []string{
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"X-CRSF-Token",
	}

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods(allowedMethods),
		handlers.AllowedHeaders(allowedHeaders),
		handlers.AllowCredentials(),
	)

	return corsMiddleware
}
