package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"

	cf "github.com/seankim658/skullking/internal/config"
	h "github.com/seankim658/skullking/internal/handlers"
	l "github.com/seankim658/skullking/internal/logger"
	mw "github.com/seankim658/skullking/internal/middleware"
	"github.com/seankim658/skullking/internal/sse"
)

func New(cfg *cf.Config, sseHub *sse.Hub) http.Handler {
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

	cachingFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		fileServer.ServeHTTP(w, r)
	})

	mainRouter.PathPrefix(avatarWebPrefix).Handler(http.StripPrefix(avatarWebPrefix, cachingFileServer))

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
	authSubRouter.HandleFunc("/initiate-link/{provider}", authHandler.HandleInitiateLink).Methods(http.MethodGet)
	authSubRouter.HandleFunc("/logout", authHandler.HandleLogout).Methods(http.MethodGet, http.MethodPost)
	authSubRouter.HandleFunc("/me", authHandler.HandleGetCurrentUser).Methods(http.MethodGet)

	// Handlers
	settingsHandler := h.NewSettingsHandler(cfg)
	gameHandler := h.NewGameHandler(cfg, sseHub)
	scoringHandler := h.NewScoringHandler(cfg, sseHub)
	sessionHandler := h.NewSessionHandler(cfg)
	userHandler := h.NewUserProfileHandler(cfg)
	friendshipHandler := h.NewFriendshipHandler(cfg, sseHub)
	notificationHandler := h.NewNotificationHandler(cfg, sseHub)
	statsHandler := h.NewStatsHandler(cfg)
	adminHandler := h.NewAdminHandler(cfg, sseHub)

	// Settings routes
	settingsSubRouter := apiRouter.PathPrefix("/settings").Subrouter()
	settingsSubRouter.Use(mw.AuthRequired)
	settingsSubRouter.HandleFunc("/theme", settingsHandler.HandleUpdateUserTheme).Methods(http.MethodPut)
	settingsSubRouter.HandleFunc("/profile", settingsHandler.HandleUpdateUserProfile).Methods(http.MethodPut)
	settingsSubRouter.HandleFunc("/linked-accounts", settingsHandler.HandleGetLinkedAccounts).Methods(http.MethodGet)
	settingsSubRouter.HandleFunc("linked-accounts/{provider}", settingsHandler.HandleUnlinkAccount).Methods(http.MethodDelete)

	// Game routes
	gameSubRouter := apiRouter.PathPrefix("/games").Subrouter()
	gameSubRouter.Use(mw.AuthRequired)
	gameSubRouter.HandleFunc("", gameHandler.HandleCreateGame).Methods(http.MethodPost)
	gameSubRouter.HandleFunc("/active", gameHandler.HandleGetActiveGames).Methods(http.MethodGet)
  gameSubRouter.HandleFunc("/pending", gameHandler.HandleGetPendingGames).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/history", gameHandler.HandleGetGameHistory).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/details", gameHandler.HandleGetGameDetails).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/players", gameHandler.HandleAddPlayerToGame).Methods(http.MethodPost)
	gameSubRouter.HandleFunc("/{game_id}/players", gameHandler.HandleGetGamePlayers).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/players/{game_player_id}", gameHandler.HandleRemovePlayerFromGame).Methods(http.MethodDelete)
	gameSubRouter.HandleFunc("/{game_id}/settings", gameHandler.HandleUpdateGameSettings).Methods(http.MethodPut)
	gameSubRouter.HandleFunc("/{game_id}/start", gameHandler.HandleStartGame).Methods(http.MethodPut)
	gameSubRouter.HandleFunc("/{game_id}/asterisks", gameHandler.HandleGetAsterisks).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/players/{game_player_id}/asterisk", gameHandler.HandleAddAsterisk).Methods(http.MethodPost)
	gameSubRouter.HandleFunc("/{game_id}/scorecard", scoringHandler.HandleGetScorecardState).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/summary", gameHandler.HandleGetGameSummary).Methods(http.MethodGet)
	gameSubRouter.HandleFunc("/{game_id}/rounds/{round_number:[0-9]+}/bids", scoringHandler.HandleSubmitBids).Methods(http.MethodPost)
	gameSubRouter.HandleFunc("/{game_id}/rounds/{round_number:[0-9]+}/tricks", scoringHandler.HandleSubmitTricks).Methods(http.MethodPost)

	// Session routes
	sessionSubRouter := apiRouter.PathPrefix("/sessions").Subrouter()
	sessionSubRouter.Use(mw.AuthRequired)
	sessionSubRouter.HandleFunc("/active", sessionHandler.HandleGetActiveSessionsForUser).Methods(http.MethodGet)
	sessionSubRouter.HandleFunc("/history", sessionHandler.HandleGetUserSessionHistory).Methods(http.MethodGet)
	sessionSubRouter.HandleFunc("/{session_id}", sessionHandler.HandleGetSessionDetails).Methods(http.MethodGet)
	sessionSubRouter.HandleFunc("/{session_id}/complete", sessionHandler.HandleCompleteSession).Methods(http.MethodPut)

	// User profile routes
	userSubRouter := apiRouter.PathPrefix("/users").Subrouter()
	userSubRouter.HandleFunc("/{user_id}/profile", userHandler.HandleGetUserProfile).Methods(http.MethodGet)
	userSubRouter.HandleFunc("/{user_id}/friends", userHandler.HandleGetFriendsList).Methods(http.MethodGet)
	userSubRouter.HandleFunc("/{user_id}/stats/awards", userHandler.HandleGetUserAwardStats).Methods(http.MethodGet)
	userSubRouter.HandleFunc("/search", userHandler.HandleSearchUsers).Methods(http.MethodGet)

	// Report routes
	reportUserSubRouter := userSubRouter.Path("/{user_id}/report").Subrouter()
	reportUserSubRouter.Use(mw.AuthRequired)
	reportUserSubRouter.HandleFunc("", userHandler.HandleReportUser).Methods(http.MethodPost)

	// Friendship routes
	friendshipSubRouter := apiRouter.PathPrefix("/friends").Subrouter()
	friendshipSubRouter.Use(mw.AuthRequired)
	friendshipSubRouter.HandleFunc("", friendshipHandler.HandleGetFriends).Methods(http.MethodGet)
	friendshipSubRouter.HandleFunc("/request", friendshipHandler.HandleSendFriendRequest).Methods(http.MethodPost)
	friendshipSubRouter.HandleFunc("/request/{friendship_id}", friendshipHandler.HandleRespondToFriendRequest).Methods(http.MethodPut)
	friendshipSubRouter.HandleFunc("/{user_id}", friendshipHandler.HandleUnfriend).Methods(http.MethodDelete)
	friendshipSubRouter.HandleFunc("/request/cancel/{addressee_id}", friendshipHandler.HandleCancelFriendRequest).Methods(http.MethodDelete)
	friendshipSubRouter.HandleFunc("/block/{user_id_to_block}", friendshipHandler.HandleBlockUser).Methods(http.MethodPost)
	friendshipSubRouter.HandleFunc("/block/{user_id_to_unblock}", friendshipHandler.HandleUnblockUser).Methods(http.MethodDelete)

	// Notification routes
	notificationSubRouter := apiRouter.PathPrefix("/notifications").Subrouter()
	notificationSubRouter.Use(mw.AuthRequired)
	notificationSubRouter.HandleFunc("", notificationHandler.HandleGetNotifications).Methods(http.MethodGet)
	notificationSubRouter.HandleFunc("", notificationHandler.HandleDeleteAllNotifications).Methods(http.MethodDelete)
	notificationSubRouter.HandleFunc("/events", notificationHandler.HandleNotificationStream).Methods(http.MethodGet, http.MethodOptions)
	notificationSubRouter.HandleFunc("/{notification_id}", notificationHandler.HandleDeleteNotification).Methods(http.MethodDelete)
	notificationSubRouter.HandleFunc("/{notification_id}/read", notificationHandler.HandleMarkNotificationRead).Methods(http.MethodPut)
	notificationSubRouter.HandleFunc("/{notification_id}/read", notificationHandler.HandleMarkNotificationUnread).Methods(http.MethodDelete)

	// Stats routes
	statsSubRouter := apiRouter.PathPrefix("/stats").Subrouter()
	statsSubRouter.HandleFunc("/summary", statsHandler.HandleGetSiteSummaryStats).Methods(http.MethodGet)
	statsSubRouter.HandleFunc("/leaderboard", statsHandler.HandleGetGlobalLeaderboard).Methods(http.MethodGet)

	// Admin routes
	adminSubrouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminSubrouter.Use(mw.AuthRequired)
	adminSubrouter.Use(mw.SuperuserRequired)
	adminSubrouter.HandleFunc("/users/{user_id}/ban", adminHandler.HandleBanUser).Methods(http.MethodPost)
	adminSubrouter.HandleFunc("/users/{user_id}/unban", adminHandler.HandleUnbanUser).Methods(http.MethodPost)
	adminSubrouter.HandleFunc("/reports", adminHandler.HandleGetReports).Methods(http.MethodGet)
	adminSubrouter.HandleFunc("/reports/{report_id}", adminHandler.HandleUpdateReportStatus).Methods(http.MethodPut)
	adminSubrouter.HandleFunc("/notifications", adminHandler.HandleSendAdminNotification).Methods(http.MethodPost)

	return mainRouter
}
