package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	i "github.com/seankim658/skullking/internal/images"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const settingsComponent = "handlers-settings"

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
		settingsComponent,
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

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update theme settings")
	if !txOk {
		return
	}

	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	params := dbModels.UpdateUserParams{
		UITheme:    &payload.UITheme,
		ColorTheme: &payload.ColorTheme,
	}

	opErr = db.UpdateUserProfile(ctx, tx, userID, params)
	if opErr != nil {
		if errors.Is(opErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User not found, cannot update theme")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update theme settings")
		}
		responseSent = true
		return
	}

	if !responseSent {
		FetchUserAndRespond(
			w, r, nil, userID, logger, http.StatusOK,
			"Theme settings updated successfully",
		)
		responseSent = true
	}
}

// Updates the user's profile information
// Path: /settings/profile
// Method: PUT
func (sh *SettingsHandler) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsComponent,
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

	params := dbModels.UpdateUserParams{
		StatsPrivacy: payload.StatsPrivacy,
	}

	if payload.DisplayName != nil {
		trimmedDisplayName := strings.TrimSpace(*payload.DisplayName)
		if trimmedDisplayName == "" {
			ErrorResponse(w, r, http.StatusBadRequest, "Display naem cannot be empty or only whitespace")
			return
		}
		params.DisplayName = &trimmedDisplayName
	}

	if payload.AvatarURL != nil {
		originalReqAvatarURL := *payload.AvatarURL
		if strings.HasPrefix(originalReqAvatarURL, "http") {
			// User provided a new external URL for manual avatar update
			logger.Info().Msg("User provided an external avatar URL for manual update")
			localPath, processErr := i.ProcessAndStoreAvatar(
				ctx, originalReqAvatarURL, userID,
				sh.Cfg.AvatarStoragePath,
				i.AvatarWebPrefixPath,
				i.AvatarImgSize,
			)
			if processErr != nil {
				ErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Failed to process the provided avatar URL: %v", processErr))
				return
			}
			params.AvatarURL = &localPath
			manualSource := i.AvatarManualKey
			params.AvatarSource = &manualSource
			logger.Info().Str(l.PathKey, localPath).Msg("Successfully localized and set manually provided avatar")
		} else if originalReqAvatarURL == "" {
			// User wants to remove their avatar
			emtpyString := ""
			params.AvatarSource = &emtpyString
			params.AvatarSource = &emtpyString
			logger.Info().Msg("User reqeusted to remove avatar")
		}
	}

	if params.DisplayName == nil && params.StatsPrivacy == nil && params.AvatarURL == nil {
		logger.Info().Msg("No updatable fields provided in payload for profile update")
		FetchUserAndRespond(w, r, nil, userID, logger, http.StatusOK, "No profile information was updated")
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update user profile")
	if !txOk {
		return
	}

	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.UpdateUserProfile(ctx, tx, userID, params)
	if opErr != nil {
		logger.Error().Err(opErr).Interface(l.UpdatesKey, params).Msg("Failed to update user profile in database")
		if errors.Is(opErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "User not found, cannot update profile")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update profile information")
		}
		responseSent = true
		return
	}

	if !responseSent {
		FetchUserAndRespond(w, r, nil, userID, logger, http.StatusOK, "Profile updated successfully")
		responseSent = true
	}
}

// Retrieves the list of OAuth accounts linked to the current user
// Path: /settings/linked-accounts
// Method: GET
func (sh *SettingsHandler) HandleGetLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		settingsComponent,
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
		settingsComponent,
		"HandleUnlinkAccount",
	)

	providerName, ok := PathVar(w, r, "provider")
	if !ok {
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.DeleteUserProviderIdentity(ctx, tx, userID, providerName)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to delete provider identity from database")
		if errors.Is(opErr, db.ErrUserProviderIdentityNotFound) {
			ErrorResponse(
				w, r, http.StatusNotFound,
				"The specified account to unlink was not found for your user",
			)
		} else if errors.Is(opErr, db.ErrDeleteLastProviderIdentity) {
			ErrorResponse(
				w, r, http.StatusBadRequest,
				"Cannot unlink the last authentication method. Please link another account first or ensure you have an alternative login method",
			)
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to unlink account")
		}
		responseSent = true
		return
	}

	if !responseSent {
		Respond(
			w, r, http.StatusOK, nil,
			fmt.Sprintf("Account with provider '%s' unlinked successfully.", providerName),
		)
		responseSent = true
	}
}
