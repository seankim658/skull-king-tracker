package auth

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"github.com/seankim658/skullking/internal/config"
	l "github.com/seankim658/skullking/internal/logger"
)

var log = l.WithComponent(l.AppLog, "auth-auth")

const (
	// Specifies how long the session cookie will be valid in seconds
	SessionMaxAge = 86400 * 30 // 30 days
	// Name of the cookie that will store the session ID
	SessionCookieName = "skullking_auth_session"
	// Key used to store the authenticated user's ID within the session data
	UserIDSessionKey = "user_id"
	// Key used to store the authenticated user's display name within the session data
	UserNameSessionKey = "user_name"
  // Key used to store the provider name during an account linking operation
  LinkingProviderNameSessionKey = "linking_provider_name"
  // Key used to store the user ID initiating an account linking operation
  LinkingUserIDSessionKey = "linking_user_id_for_provider"
)

// Initializes the goth authentication providers, should be called once at applciation startup
func InitAuth(cfg *config.Config) error {
	logger := l.WithSource(log, "InitAuth")
	logger.Info().Msg("Initializing authentication systems...")

	// --- Session Store Initialization ---
	store := sessions.NewCookieStore([]byte(cfg.SessionSecretKey), []byte(cfg.SessionEncryptionKey))
	store.MaxAge(SessionMaxAge)   // How long the browser will keeps the session cookie
	store.Options.Path = "/"      // Cookie is valid for all paths on the domain
	store.Options.HttpOnly = true // Prevents client-side JavaScript from accessing the cookie
	store.Options.Secure = cfg.AppEnv == "production"
  if cfg.AppEnv == "production" {
    store.Options.Secure = true
    store.Options.SameSite = http.SameSiteLaxMode
  } else {
    store.Options.Secure = false
    store.Options.SameSite = http.SameSiteLaxMode
  }

	gothic.Store = store
	logger.Info().
		Str("cookie_name", SessionCookieName).
		Int("max_age_seconds", SessionMaxAge).
		Bool("http_only", store.Options.HttpOnly).
		Bool("secure_cookies", store.Options.Secure).
		Msg("Session store initialized and configured for gothic")

	gob.Register(map[string]any{})
	gob.Register(goth.User{})

	// --- Goth Provider Initialization ---
	var activeProviders []goth.Provider
	var initErrors []error

	// Google Provider
	googleClientID := cfg.ProviderAuthConfig.GoogleClientID
	googleClientSecret := cfg.ProviderAuthConfig.GoogleClientSecret

	if googleClientID != "" || googleClientSecret != "" {
		if googleClientID == "" {
			initErrors = append(initErrors, errors.New("GOOGLE_CLIENT_ID is set, but GOOGLE_CLIENT_SECRET is missing"))
		}
		if googleClientSecret == "" {
			initErrors = append(initErrors, errors.New("GOOGLE_CLIENT_SECRET is set, but GOOGLE_CLIENT_ID is missing"))
		}

		if googleClientID != "" && googleClientSecret != "" {
			googleCallbackURL := fmt.Sprintf("%s/api/auth/google/callback", cfg.AppBaseURL)

			logger.Info().
				Str(l.ProviderKey, "google").
				Str("callback_url", googleCallbackURL).
				Msg("Configuring Google OAuth")

			activeProviders = append(
				activeProviders,
				google.New(
					googleClientID,
					googleClientSecret,
					googleCallbackURL,
					"email",
					"profile",
				),
			)
		}
	}

	if len(initErrors) > 0 {
		for _, err := range initErrors {
			logger.Error().Err(err).Msg("OAuth provider configuration error")
		}
		return fmt.Errorf("found %d error(s) in OAuth provider configuration", len(initErrors))
	}

	if len(activeProviders) > 0 {
		goth.UseProviders(activeProviders...)
		logger.Info().Msgf("Initialized %d authentication provider(s)", len(activeProviders))
	} else {
		logger.Fatal().Msg("No authentication providers were configured or enabled")
	}

	return nil
}
