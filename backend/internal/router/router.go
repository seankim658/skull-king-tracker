package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"

	cf "github.com/seankim658/skullking/internal/config"
	h "github.com/seankim658/skullking/internal/handlers"
	l "github.com/seankim658/skullking/internal/logger"
	mw "github.com/seankim658/skullking/internal/middleware"
)

func New(cfg *cf.Config) http.Handler {
	mainRouter := mux.NewRouter()

	// --- Middleware Setup ---

	// Apply logging middlewre
	mainRouter.Use(mw.RecoveryMiddleware(l.AppLog))
	mainRouter.Use(mw.LoggingMiddleware(l.AccessLog, l.AppLog))
	mainRouter.Use(mw.CorsMiddleware(cfg))
	mainRouter.Use(mw.RateLimit(rate.Limit(5), 10))

	// --- API Subrouter ---
	apiRouter := mainRouter.PathPrefix("/api").Subrouter()

	// --- Route Definitions ---

	// Health check endpoint
	apiRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Auth routes
	authHandler := h.NewAuthHandler(cfg)
	authSubRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authSubRouter.HandleFunc("/{provider}/login", authHandler.HandleOAuthLogin).Methods(http.MethodGet)
	authSubRouter.HandleFunc("/{provider}/callback", authHandler.HandleOAuthCallback).Methods(http.MethodGet)
	authSubRouter.HandleFunc("/logout", authHandler.HandleLogout).Methods(http.MethodGet, http.MethodPost)
	authSubRouter.HandleFunc("/me", authHandler.HandleGetCurrentUser).Methods(http.MethodGet)

	// Settings routes
	settingsHandler := h.NewSettingsHandler(cfg)
	settingsSubRouter := apiRouter.PathPrefix("/settings").Subrouter()
	settingsSubRouter.HandleFunc("/theme", settingsHandler.HandleUpdateUserTheme).Methods(http.MethodPut)
	settingsSubRouter.HandleFunc("/profile", settingsHandler.HandleUpdateUserProfile).Methods(http.MethodPut)
	settingsSubRouter.HandleFunc("/linked-accounts", settingsHandler.HandleGetLinkedAccounts).Methods(http.MethodGet)
	settingsSubRouter.HandleFunc("linked-accounts/{provider}", settingsHandler.HandleUnlinkAccount).Methods(http.MethodDelete)

	return mainRouter
}
