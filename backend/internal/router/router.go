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

	// --- Static Files ---
	avatarWebPrefix := "/static/avatars/"
	avatarDiskPath := cfg.AvatarStoragePath
	fileServer := http.FileServer(http.Dir(avatarDiskPath))

	mainRouter.PathPrefix(avatarWebPrefix).Handler(http.StripPrefix(avatarWebPrefix, fileServer))

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

	// Game routes
	gameHandler := h.NewGameHandler(cfg)
	gameSubRouter := apiRouter.PathPrefix("/games").Subrouter()
	gameSubRouter.HandleFunc("", gameHandler.HandleCreateGame).Methods(http.MethodPost)
	gameSubRouter.HandleFunc("/{game_id}/players", gameHandler.HandleAddPlayerToGame).Methods(http.MethodPost)

	// Session routes
	sessionHandler := h.NewSessionHandler(cfg)
	sessionSubRouter := apiRouter.PathPrefix("/sessions").Subrouter()
	sessionSubRouter.HandleFunc("/active", sessionHandler.HandleGetActiveSessionsForUser).Methods(http.MethodGet)
	sessionSubRouter.HandleFunc("/{session_id}/complete", sessionHandler.HandleCompleteSession).Methods(http.MethodPut)

	// User prfile routes
	userProfileHandler := h.NewUserProfileHandler(cfg)
	userProfileSubRouter := apiRouter.PathPrefix("/users").Subrouter()
	userProfileSubRouter.HandleFunc("/{user_id}/profile", userProfileHandler.HandleGetUserProfile).Methods(http.MethodGet)

	return mainRouter
}
