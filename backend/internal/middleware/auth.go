package middleware

import (
	"net/http"

	db "github.com/seankim658/skullking/internal/database"
	h "github.com/seankim658/skullking/internal/handlers"
	l "github.com/seankim658/skullking/internal/logger"
)

// Ensures a user is authenticated
func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := l.GetLoggerFromContext(ctx)

		userID, ok := h.GetAuthenticatedUserIDFromSession(w, r, logger)
		if !ok {
			return
		}

		user, err := db.GetUserByID(ctx, nil, userID)
		if err != nil {
			logger.Error().Err(err).Str(l.UserIDKey, userID).Msg("AuthRequired: Failed to retrieve user")
			h.ErrorResponse(w, r, http.StatusInternalServerError, "Error verifying user session")
			return
		}

		if user.IsBanned {
			logger.Warn().Str(l.UserIDKey, userID).Msg("AuthRequired: Blocked access for banned user")
			h.ErrorResponse(w, r, http.StatusForbidden, "This account has been suspended")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Ensures that authenticated user has the 'superuser' role
func SuperuserRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := l.GetLoggerFromContext(ctx)

		userID, ok := h.GetAuthenticatedUserIDFromSession(w, r, log)
		if !ok {
			return
		}

		user, err := db.GetUserByID(ctx, nil, userID)
		if err != nil {
			log.Error().Err(err).Str(l.UserIDKey, userID).Msg("Failed to retrieve user for superuser check")
			h.ErrorResponse(w, r, http.StatusInternalServerError, "Error verifying user permissions")
			return
		}

		if user.Role != "superuser" {
			log.Warn().Str(l.UserIDKey, userID).Str(l.UserRoleKey, user.Role).Msg("Forbidden: User is not a superuser")
			h.ErrorResponse(w, r, http.StatusForbidden, "Forbidden: You do not have permission to access this resource")
			return
		}

		next.ServeHTTP(w, r)
	})
}
