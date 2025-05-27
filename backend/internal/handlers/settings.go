package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gorilla/mux"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
)

const settingsHandlerComponent = "handlers-settings"

type SettingsHandler struct {
	Cfg *cf.Config
}

func NewSettingsHandler(cfg *cf.Config) *SettingsHandler {
	return &SettingsHandler{Cfg: cfg}
}

// Updates the user's theme preferences
// Path: /settings/theme
// Method: PUT
func (sh *SettingsHandler) HandleUpdateUserTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsHandlerComponent,
		"HandleUpdateUserTheme",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	var payload apiModels.UpdateUserThemeRequest
	if !ParseJSON(w, r, &payload) {
		return
	}

	if !RequireFields(w, r, map[string]string{
		"ui_theme":    payload.UITheme,
		"color_theme": payload.ColorTheme,
	}) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update theme settings")
	if !txOk {
		return
	}

	var opErr error
	defer func() {
		if p := recover(); p != nil {
			logger.Error().
				Interface(l.PanicKey, p).
				Bytes(l.StackTraceKey, debug.Stack()).
				Msg("Panic recovered during theme update")
			_ = tx.Rollback()
		} else if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction due to error in handler logic")
			rbErr := tx.Rollback()
			if rbErr != nil {
				logger.Error().Err(rbErr).Msg("Transaction rollback failed after handler error")
			}
		}
	}()

	opErr = db.UpdateUserThemeSettings(ctx, tx, userID, payload.UITheme, payload.ColorTheme)
	if opErr != nil {
		if errors.Is(opErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User not found, cannot update theme")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update theme settings")
		}
		return
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		opErr = commitErr
		logger.Error().Err(opErr).Msg("Transaction commit failed for theme update")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save theme settings")
		return
	}
	logger.Debug().Msg("Transaction committed successfully for theem update")

	FetchUserAndRespond(
		w,
		r,
		nil,
		userID,
		logger,
		http.StatusOK,
		"Theme settings updated successfully",
	)
}

// Updates the user's profile information
// Path: /settings/profile
// Method: PUT
func (sh *SettingsHandler) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsHandlerComponent,
		"HandleUpdateUserProfile",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	var payload apiModels.UpdateUserProfileRequest
	if !ParseJSON(w, r, &payload) {
		return
	}

	updates := make(map[string]any)
	if payload.DisplayName != nil {
		trimmedDisplayName := strings.TrimSpace(*payload.DisplayName)
		if trimmedDisplayName == "" {
			ErrorResponse(w, r, http.StatusBadRequest, "Display name cannot be empty or only whitespace")
			return
		}
		updates["display_name"] = trimmedDisplayName
	}
	if payload.AvatarURL != nil {
		updates["avatar_url"] = *payload.AvatarURL
	}
	if payload.StatsPrivacy != nil {
		updates["stats_privacy"] = *payload.StatsPrivacy
	}

	if len(updates) == 0 {
		logger.Info().Msg("No updatable fields provided in payload for profile update")
		FetchUserAndRespond(w, r, nil, userID, logger, http.StatusOK, "No profile information was updated")
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update user profile")
	if !txOk {
		return
	}

	var opErr error
	defer func() {
		if p := recover(); p != nil {
			logger.Error().Interface(l.PanicKey, p).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic recovered during profile update")
			_ = tx.Rollback()
		} else if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction due to error in profile update handler logic")
			_ = tx.Rollback()
		}
	}()

	opErr = db.UpdateUserProfile(ctx, tx, userID, updates)
	if opErr != nil {
		logger.Error().Err(opErr).Interface(l.UpdatesKey, updates).Msg("Failed to update user profile in database")
		if errors.Is(opErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User not found, cannot update profile")
		} else if errors.Is(opErr, db.ErrInvalidStatsPrivacy) {
			ErrorResponse(w, r, http.StatusBadRequest, opErr.Error())
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update profile information")
		}
		return
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		opErr = commitErr
		logger.Error().Err(opErr).Msg("Transaction commit failed for profile update")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save profile changes")
		return
	}
	logger.Info().Msg("Transaction for profile update committed successfully")

	FetchUserAndRespond(w, r, nil, userID, logger, http.StatusOK, "Profile updated successfully")
}

// Retrieves the list of OAuth accounts linked to the current user
// Path: /settings/linked-accounts
// Method: GET
func (sh *SettingsHandler) HandleGetLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsHandlerComponent,
		"HandleGetLinkedAccounts",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	dbIdentities, dbErr := db.GetUserProviderIdentitiesByUserID(ctx, nil, userID)
	if dbErr != nil {
		logger.Error().Err(dbErr).Msg("Failed to retrieve linked accounts from database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrive linked accounts")
		return
	}

	apiLinkedAccounts := make([]apiModels.LinkedAccount, 0, len(dbIdentities))
	for _, dbID := range dbIdentities {
		linkedAccount, convErr := modelConverters.DBProviderIdentityToLinkedAccount(&dbID)
		if convErr != nil {
			logger.Error().
				Err(convErr).
				Msg("Failed to convert from provider identity to linked account")
			ErrorResponse(
				w, r,
				http.StatusInternalServerError,
				"failed to convert from provider identity to linked account",
			)
			return
		}
		apiLinkedAccounts = append(apiLinkedAccounts, *linkedAccount)
	}

	Respond(w, r, http.StatusOK, apiLinkedAccounts, "successfully retrieved linked accounts")
}

// Allows a user to unlink one of their OAuth provider accounts
// Path: /settings/linked-accounts/{provider}
// Method: DELETE
func (sh *SettingsHandler) HandleUnlinkAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsHandlerComponent,
		"HandleUnlinkAccount",
	)

	vars := mux.Vars(r)
	providerName, ok := vars["provider"]
	if !ok || providerName == "" {
		ErrorResponse(w, r, http.StatusBadRequest, "Provider name is required in the URL path")
		return
	}
	logger = logger.With().Str(l.ProviderKey, providerName).Logger()

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to process unlink request.")
	if !txOk {
		return
	}

	var opErr error
	defer func() {
		if p := recover(); p != nil {
			logger.Error().
				Interface(l.PanicKey, p).
				Bytes(l.StackTraceKey, debug.Stack()).
				Msg("Panic recovered during unlink account")
			_ = tx.Rollback()
		} else if opErr != nil {
			logger.Warn().
				Err(opErr).
				Msg("Rolling back transaction due to error in unlink account handler logic")
			_ = tx.Rollback()
		}
	}()

	opErr = db.DeleteUserProviderIdentity(ctx, tx, userID, providerName)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to delete provider identity from database")
		if errors.Is(opErr, db.ErrUserProviderIdentityNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "The specified account to unlink was not found for your user.")
		} else if errors.Is(opErr, db.ErrDeleteLastProviderIdentity) {
			ErrorResponse(w, r, http.StatusBadRequest, "Cannot unlink the last authentication method. Please link another account first or ensure you have an alternative login method.")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to unlink account.")
		}
		return
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		opErr = commitErr
		logger.Error().Err(opErr).Msg("Transaction commit failed for unlinking account")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize unlinking account.")
		return
	}
	logger.Info().Msg("Transaction for unlinking account committed successfully.")

	Respond(
		w, r,
		http.StatusOK,
		nil,
		fmt.Sprintf("Account with provider '%s' unlinked successfully.", providerName),
	)
}
