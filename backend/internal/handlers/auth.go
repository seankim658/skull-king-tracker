package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"

	a "github.com/seankim658/skullking/internal/auth"
	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	dbModels "github.com/seankim658/skullking/internal/models/database"
	u "github.com/seankim658/skullking/internal/users"
)

const authComponent = "handlers-auth"

// Handles HTTP requests for user authentication
type AuthHandler struct {
	Cfg *cf.Config
}

// Creates a new `AuthHandler`
func NewAuthHandler(cfg *cf.Config) *AuthHandler {
	return &AuthHandler{
		Cfg: cfg,
	}
}

// Initiates the OAuth2 authentication flow for a given provider (redirects the user to the provider's OAuth page)
// Path: /auth/{proivder}/login
// Method: GET
func (ah *AuthHandler) HandleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(r.Context()),
		authComponent,
		"HandleOAuthLogin",
	)

	vars := mux.Vars(r)
	providerName, ok := vars["provider"]
	if !ok || providerName == "" {
		logger.Error().Msg("Provider name missing from URL path")
		ErrorResponse(w, r, http.StatusBadRequest, "Authentication proivder name is required")
		return
	}

	logger.Info().Str(l.ProviderKey, providerName).Msg("Attempting OAuth login")

	// Add provider name to the request context for gothic
	ctxWithProvider := context.WithValue(ctx, gothic.ProviderParamKey, providerName)
	rWithProvider := r.WithContext(ctxWithProvider)

	gothic.BeginAuthHandler(w, rWithProvider)
}

// Initiates the OAuth2 flow for linking an additional provider to an existing authenticated user
// Path: /auth/initiate-link/{provider}
// Method: GET
func (ah *AuthHandler) HandleInitiateLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		authComponent,
		"HandleInitiateLink",
	)

	// Step 1: Ensure user is authenticated
	sessionUserID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, sessionUserID).Logger()

	// Step 2: Get provider from URL
	vars := mux.Vars(r)
	providerName, foundProvider := vars["provider"]
	if !foundProvider || providerName == "" {
		logger.Error().Msg("Provider name missing from URL path for linking")
		ErrorResponse(w, r, http.StatusBadRequest, "Provider name is required for linking")
		return
	}
	logger = logger.With().Str(l.ProviderKey, providerName).Logger()

	// Step 3: Set session flags for linking
	session, err := gothic.Store.Get(r, a.SessionCookieName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get session store for linking")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to initiate account linking: session error")
		return
	}

	session.Values[a.LinkingUserIDSessionKey] = sessionUserID
	session.Values[a.LinkingProviderNameSessionKey] = providerName

	if err := session.Save(r, w); err != nil {
		logger.Error().Err(err).Msg("Failed to save session for linking flag")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to initiate account linking: session save error")
		return
	}
	logger.Debug().Msg("Session flags set for linking, beginning OAuth flow with provider")

	// Step 4: Begin OAuth flow
	ctxWithProvider := context.WithValue(ctx, gothic.ProviderParamKey, providerName)
	rWithProvider := r.WithContext(ctxWithProvider)
	gothic.BeginAuthHandler(w, rWithProvider)
}

// Endpoint where the OAuth provider redirects the user after they have authenticated (or denied) access on the provider site.
// Function is responsible for:
// 1. Completing the OAuth flow with the provider to get user details
// 2. Determining if this is part of an "account linking" flow or a login flow
// 3. Performing database operations within a transaction:
//   - For linking: associating the new provider identity with the existing logged-in user
//   - For login: finding the user by this provider identity
//   - For registration: creating a new user if no existing account if found (by provider ID or email)
//
// 4. Handling potential conflicts (e.g., provider account already linked to another user)
// 5. Updating user's last login time
// 6. Establishing or updating the application session
// 7. Redirecting the user to the appropriate frontend page
// Path: /auth/{provider}/callback
// Method: GET
func (ah *AuthHandler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(r.Context()),
		authComponent,
		"HandleOAuthCallback",
	)

	vars := mux.Vars(r)
	providerName := vars["provider"]

	logger.Info().Str(l.ProviderKey, providerName).Msg("Received OAuth callback")

	// --- Step 1 ---
	// Exchange the code from the provider for an access token and fetch user details

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		logger.Error().Err(err).Str(l.ProviderKey, providerName).Msg("Failed to complete OAuth user authentication")
		ErrorResponse(w, r, http.StatusInternalServerError, "Authentication failed: "+err.Error())
		return
	}

	// Ensure gothUser.Provider matches providerName from URL
	if gothUser.Provider != providerName {
		logger.Error().
			Str("url_provider", providerName).
			Str("goth_provider", gothUser.Provider).
			Msg("Provider mismatch between URL and Goth user details")
		ErrorResponse(w, r, http.StatusBadRequest, "Provider mismatch during authentication callback")
		return
	}

	logger.Info().
		Str(l.ProviderKey, gothUser.Provider).
		Str(l.EmailKey, gothUser.Email).
		Str(l.ProviderUserIDKey, gothUser.UserID).
		Str(l.GothUsernameKey, gothUser.Name).
		Str(l.GothNickNameKey, gothUser.NickName).
		Msg("User successfully authenticated with provider")

	var appUserID string                  // UUID of the user in the user table
	var dbUserForSession *dbModels.User   // The full user model for session data
	var responseSent bool                 // Flag to prevent sending multiple HTTP error responses
	var handleErr error                   // Track errors within main logic; checked by defer function
	redirectURL := ah.Cfg.FrontendBaseURL // Default redirect

	// --- Step 2 ---
	// Determine intent - linking existing account or login/registration

	currentSession, sessionErr := gothic.Store.Get(r, a.SessionCookieName)
	if sessionErr != nil {
		logger.Warn().Err(sessionErr).Msg("Failed to get session store when checking for linking intent")
	}

	linkingUserIDVal := currentSession.Values[a.LinkingUserIDSessionKey]
	linkingProviderVal := currentSession.Values[a.LinkingProviderNameSessionKey]
	isLinkingFlow := false
	var linkingAppUserID, linkingProviderName string

	if linkingUserIDVal != nil && linkingProviderVal != nil {
		var castOk bool
		linkingAppUserID, castOk = linkingUserIDVal.(string)
		if !castOk {
			logger.Error().Interface(l.ValueKey, linkingUserIDVal).Msg("Failed to cast linking_user_id to string")
		}
		linkingProviderName, castOk = linkingProviderVal.(string)
		if !castOk {
			logger.Error().Interface(l.ValueKey, linkingProviderVal).Msg("Failed to cast linking_provider_name to string")
		}

		// If session flags are valid and match the current provider, it's an account linking flow
		if linkingAppUserID != "" && linkingProviderName != "" && linkingProviderName == gothUser.Provider {
			isLinkingFlow = true
			logger.Info().
				Str(l.UserIDKey, linkingAppUserID).
				Str(l.ProviderKey, linkingProviderName).
				Msg("Account linking flow detected")
			appUserID = linkingAppUserID
			redirectURL = ah.Cfg.FrontendBaseURL + "/settings"

			// Clear linking flags from session to prevent reuse or stale state
			delete(currentSession.Values, a.LinkingUserIDSessionKey)
			delete(currentSession.Values, a.LinkingProviderNameSessionKey)
			if sErr := currentSession.Save(r, w); sErr != nil {
				logger.Error().
					Err(sErr).
					Msg("Failed to save session after clearing linking flags (pre-transaction)")
			}
		} else {
			// Linking flags were present but invalid
			logger.Warn().
				Str("session_linking_uid", linkingAppUserID).
				Str("session_linking_provider", linkingProviderName).
				Str("goth_provider", gothUser.Provider).
				Msg("Linking flags found in session but mismatched or incomplete; proceeding with normal login")
			delete(currentSession.Values, a.LinkingUserIDSessionKey)
			delete(currentSession.Values, a.LinkingProviderNameSessionKey)
			if sErr := currentSession.Save(r, w); sErr != nil {
				logger.Error().
					Err(sErr).
					Msg("Failed to save session after clearing stale/mismatched linking flags")
			}
		}
	}

	// --- Step 3 ---
	// All database operations occur within a single transaction

	tx, txErr := db.DB.BeginTx(ctx, nil)
	if txErr != nil {
		handleErr = fmt.Errorf("failed to begin database transaction: %w", txErr)
		logger.Error().Err(handleErr).Msg("Database transaction could not be started")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process login")
		responseSent = true
		return
	}

	defer func() {
		if p := recover(); p != nil { // Case 1: A panic occurred
			logger.Error().
				Interface(l.PanicKey, p).
				Bytes(l.StackTraceKey, debug.Stack()).
				Msg("Panic recovered during OAuth callback")

			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Error().Err(rbErr).Msg("Transaction rollback failed after panic")
			} else {
				logger.Warn().Msg("Transaction rolled back successfully due to panic")
			}

			if !responseSent {
				ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
			}

		} else if handleErr != nil { // Case 2: A known error (`handleErr`) occurred in the handler logic
			logger.Warn().Err(handleErr).Msg("Rolling back transaction due to error in handler logic")
			if tx != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					logger.Error().Err(rbErr).Msg("Transaction rollback failed after handler error")
				} else {
					logger.Info().Msg("Transaction rollback successfully after handler error")
				}
			} else {
				logger.Info().Msg("Transaction was not initiated successfully, no rollback needed")
			}

			// Error should have been called by the code that set handleErr, handling defensively
			if !responseSent {
				logger.Error().
					Msg("Inconsistency: handleError set but no response was sent prior to defer rollback, sending generic error")
				ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
			}

		} else { // Case 3: No panic and no error; proceed to commit
			logger.Debug().Msg("Attempting to commit transaction")
			if commitErr := tx.Commit(); commitErr != nil {
				handleErr = commitErr
				logger.Error().Err(handleErr).Msg("Transaction commit failed")
				if !responseSent {
					ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
				}
			} else {
				logger.Debug().Msg("Transaction committed successfully")
			}
		}
	}()

	// --- Step 4 ---
	// Find/create user and link provider identity

	// 4a. Check if this specific provider identity already exists in the database
	existingIdentity, idErr := db.GetUserProviderIdentity(ctx, tx, gothUser.Provider, gothUser.UserID)

	if isLinkingFlow {
		// --- Branch 4.1: Account Linking Flow ---
		// The user is already logged in and is trying to link an additional OAuth provider
		// `appUserID` is already set from `linkingAppUserID`
		logger.Debug().Str(l.UserIDKey, appUserID).Msg("Processing account linking flow for existing user")

		if idErr == nil { // An existing provider identity was found for this gothUser
			if existingIdentity.UserID == appUserID {
				// Case 4.1.1: This provider account is already linked to the currently logged in user
				// Treaing as a success
				logger.Info().
					Str(l.ProviderIdentityIDKey, existingIdentity.ProviderIdentityID).
					Msg("This provider account is already linked to the current user")
				if errUpdate := db.UpdateUserProviderIdentityDetails(ctx, tx, existingIdentity.ProviderIdentityID, gothUser); errUpdate != nil {
					handleErr = fmt.Errorf("failed to update existing provider identity details during re-link: %w", errUpdate)
					logger.Error().Err(handleErr).Msg("Error updating provider identity details")
					ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update linked account details")
					responseSent = true
					return
				}
			} else {
				// case 4.1.2: Conflict, this provider is already linked to a different skull king user
				handleErr = fmt.Errorf(
					"provider identity conflict during linking: provider %s (ID: %s) is already linked to a different user (ID: %s), current user is (ID: %s)",
					gothUser.Provider, gothUser.UserID, existingIdentity.UserID, appUserID,
				)
				logger.Error().
					Err(handleErr).
					Msg("Provider identity conflict: Cannot link provider account as it's already associated with another user")
				ErrorResponse(
					w, r, http.StatusConflict,
					fmt.Sprintf("This %s account is already associated with a different account", gothUser.Provider),
				)
				responseSent = true
				return
			}
		} else if errors.Is(idErr, db.ErrUserProviderIdentityNotFound) {
			// Case 4.1.3: This provider account is not yet linked to any user (ideal scenario)
			logger.Info().
				Msg("Provider identity not found in DB, creating new provider identity and linking to current user")
			newDbIdentity := &dbModels.UserProviderIdentity{
				UserID:              appUserID,
				ProviderName:        gothUser.Provider,
				ProviderUserID:      gothUser.UserID,
				ProviderEmail:       db.NullString(gothUser.Email),
				ProviderDisplayName: db.NullString(gothUser.Name),
				ProviderAvatarURL:   db.NullString(gothUser.AvatarURL),
			}
			_, createLPIErr := db.CreateUserProviderIdentity(ctx, tx, newDbIdentity)
			if createLPIErr != nil {
				handleErr = fmt.Errorf("failed to create user provider identity during linking: %w", createLPIErr)
				logger.Error().Err(handleErr).Msg("Databsae error creating provider identity for linking")
				// This should have been caught in `GetUserProviderIdentity`, but handle defensively
				if errors.Is(createLPIErr, db.ErrProviderIdentityConflict) {
					ErrorResponse(w, r, http.StatusConflict, "Failed to link account due to a conflict")
				} else {
					ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
				}
				responseSent = true
				return
			}
			logger.Info().Str(l.ProviderKey, gothUser.Provider).Msg("New provider identity created and linked to current user successfully")
		} else {
			// Case 4.1.4: An unexpected database error occurred while checking for the provider identity
			handleErr = fmt.Errorf("unexpected database error while checking provider identity during linking: %w", idErr)
			logger.Error().Err(handleErr).Msg("Unexpected Database error")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process account linking due to a database error, please try again")
			responseSent = true
			return
		}

		// After successful link or update in the linking flow, fetch the user model for session data
		dbUserForSession, handleErr = db.GetUserByID(ctx, tx, appUserID)
		if handleErr != nil {
			logger.Error().Err(handleErr).Str(l.UserIDKey, appUserID).Msg("Failed to retrieve user (for session) after linking provider")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize the linking process, please try again")
			responseSent = true
			return
		}
	} else {
		// --- Branch 4.2: Normal Login / Registration Flow  ---
		logger.Debug().Msg("Processing normal login/registration flow for provider identity")

		if idErr != nil && !errors.Is(idErr, db.ErrUserProviderIdentityNotFound) {
			// Case 4.2.1: An unexpected database error occurred while checking for the provider identity
			handleErr = fmt.Errorf("error checking for existing provider identity during login/registration: %w", idErr)
			logger.Error().Err(handleErr).Msg("Unexpected database error")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process login, please try again")
			responseSent = true
			return
		}

		if errors.Is(idErr, db.ErrUserProviderIdentityNotFound) {
			// Case 4.2.2: This provider identity is new (user has not logged in/registered with this specific provider account before)
			// Need to either:
			//  - Link to an existing user if one exists with the same email
			//  - Create a new user if no email match is found (or if email is not provided)
			logger.Debug().Msg("Provider identity not found, will check for existing user by email from provider, or create a new user")
			var existingDBUserByEmail *dbModels.User

			// 4.2.2.a: If the OAuth provider returned an email, check if a user already exists with that user
			if strings.TrimSpace(gothUser.Email) != "" {
				existingDBUserByEmail, handleErr = db.GetUserByEmail(ctx, tx, gothUser.Email)
				if handleErr != nil && !errors.Is(handleErr, db.ErrUserNotFound) {
					handleErr = fmt.Errorf("error checking for existing user by email: %w", handleErr)
					logger.Error().Err(handleErr).Msg("Unexpected database error checking for user by email")
					ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process login, please try again")
					responseSent = true
					return
				}
				if errors.Is(handleErr, db.ErrUserNotFound) {
					handleErr = nil
				}
			}

			if existingDBUserByEmail != nil {
				// Case 4.2.2.a.1: An existing user was found with this email
				// Link the new provider identity to the existing user
				logger.Info().
					Str(l.UserIDKey, existingDBUserByEmail.UserID).
					Str(l.EmailKey, gothUser.Email).Str(l.ProviderKey, gothUser.Provider).
					Msg("Existing user found by email from provider, linking new provider identity to this user")
				appUserID = existingDBUserByEmail.UserID
				dbUserForSession = existingDBUserByEmail
			} else {
				// Case 4.2.2.a.2: No existing user found with the provider's email (or the provider did not return an email)
				logger.Debug().Msg("No existing user found by email from provider (or email was empty/not provided), creating new user")

				// Prepare a default username for the user
				baseUsername := gothUser.NickName
				if baseUsername == "" {
					baseUsername = gothUser.Name
				}
				generatedUsername, genUserErr := u.GenerateUniqueUsername(ctx, baseUsername, gothUser.Email)
				if genUserErr != nil {
					handleErr = fmt.Errorf("failed to generate username for new user: %w", genUserErr)
					logger.Error().Err(handleErr).Msg("Username generation failed")
					ErrorResponse(w, r, http.StatusInternalServerError, "Failed to prepare new user account, please try again")
					responseSent = true
					return
				}

				// Create the new user model to be inserted into the database
				newUser := &dbModels.User{
					Username:    generatedUsername,
					Email:       db.NullString(gothUser.Email),
					DisplayName: db.NullString(gothUser.Name),
					AvatarURL:   db.NullString(gothUser.AvatarURL),
				}

				// Save the new user to the database
				createdUserID, createdUserErr := db.CreateUser(ctx, tx, newUser)
				if createdUserErr != nil {
					handleErr = fmt.Errorf("failed to create new user in database: %w", createdUserErr)
					logger.Error().Err(handleErr).Interface("new_user", newUser).Msg("Database errror during new user creation")
					if errors.Is(createdUserErr, db.ErrUsernameTaken) || errors.Is(createdUserErr, db.ErrEmailTaken) {
						ErrorResponse(w, r, http.StatusConflict, "This username or email is already in use")
					} else {
						ErrorResponse(w, r, http.StatusInternalServerError, "Failed to create your account due to a database error")
					}
					responseSent = true
					return
				}
				appUserID = createdUserID

				dbUserForSession, handleErr = db.GetUserByID(ctx, tx, appUserID)
				if handleErr != nil {
					handleErr = fmt.Errorf("failed to retrieve newly created user from database (ID: %s): %w", appUserID, handleErr)
					logger.Error().Err(handleErr).Msg("Critical: Newly created user could not be fetched")
					ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
					responseSent = true
					return
				}
				logger.Info().Str(l.UserIDKey, appUserID).Str(l.UsernameKey, dbUserForSession.Username).Msg("New user created successfully")
			}

			// 4.2.2.b: Create and save the new provider identity, linking it to the `appUserID`
			newDbIdentity := &dbModels.UserProviderIdentity{
				UserID:              appUserID,
				ProviderName:        gothUser.Provider,
				ProviderUserID:      gothUser.UserID,
				ProviderEmail:       db.NullString(gothUser.Email),
				ProviderDisplayName: db.NullString(gothUser.Name),
				ProviderAvatarURL:   db.NullString(gothUser.AvatarURL),
			}
			_, createLPIErr := db.CreateUserProviderIdentity(ctx, tx, newDbIdentity)
			if createLPIErr != nil {
				handleErr = fmt.Errorf("failed to create user provider identity record: %w", createLPIErr)
				logger.Error().Err(handleErr).Interface("new_db_identity", newDbIdentity).Msg("Database error creating user provider identity record")
				if errors.Is(createLPIErr, db.ErrProviderIdentityConflict) {
					// Shouldn't happen, handling defensively
					ErrorResponse(w, r, http.StatusConflict, "Failed to link authentication method due to a conflict")
				} else {
					ErrorResponse(w, r, http.StatusInternalServerError, "Failed to link authentication method")
				}
				responseSent = true
				return
			}
			logger.Info().
				Str(l.UserIDKey, appUserID).
				Str(l.ProviderKey, gothUser.Provider).
				Msg("New user provider identity created and linked successfully")
		} else {
			// Case 4.2.3: Provider identity was found, standard login scenario
			logger.Debug().Str(l.ProviderIdentityIDKey, existingIdentity.ProviderIdentityID).Msg("Existing provider identity found")
			appUserID = existingIdentity.UserID

			dbUserForSession, handleErr = db.GetUserByID(ctx, tx, appUserID)
			if handleErr != nil {
				handleErr = fmt.Errorf("failed to retrieve user by ID (ID: %s) for existing provider identity: %w", appUserID, handleErr)
				logger.Error().Err(handleErr).Msg("Database error fetching user with existing provider identity")
				ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process your login, please try again")
				responseSent = true
				return
			}

			if errUpdate := db.UpdateUserProviderIdentityDetails(ctx, tx, existingIdentity.ProviderIdentityID, gothUser); errUpdate != nil {
				logger.Warn().
					Err(errUpdate).
					Str(l.ProviderIdentityIDKey, existingIdentity.ProviderIdentityID).
					Msg("Could not update provider details upon login, proceeding with login")
			}
		}
	} // End of `isLinkingFlow` vs. Normal Login/Registration Flow

	// Step 4.X: Sanity Check - Ensure `dbUserForSession` is populated
	if dbUserForSession == nil {
		if handleErr == nil {
			handleErr = errors.New("CRITICAL: dbUserForSession is nil, but no preceeding error was recorded")
		}
		logger.Error().Err(handleErr).Msg("User data for session (dbUserForSession) was not populated")
		if !responseSent {
			ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
			responseSent = true
		}
		return
	}

	// --- Step 5 ---
	// Update user's last login timestamp
	updateLoginTimeErr := db.UpdateUserLastLogin(ctx, tx, appUserID)
	if updateLoginTimeErr != nil {
		logger.Error().Err(updateLoginTimeErr).Str(l.UserIDKey, appUserID).Msg("Failed to update user's last login time, proceeding")
	} else {
		logger.Debug().Str(l.UserIDKey, appUserID).Msg("User's last login time updated successfully")
	}

	if handleErr != nil {
		logger.Warn().Err(handleErr).Msg("Skipping session creation due to earlier critical error")
		return
	}

	// --- Step 6 ---
	// Create or update application session
	finalSession, finalSessionErr := gothic.Store.Get(r, a.SessionCookieName)
	if finalSessionErr != nil {
		handleErr = fmt.Errorf("failed to get or create final session store for session setup: %w", finalSessionErr)
		logger.Error().Err(handleErr).Msg("Critical: Failed to get session store for final session setup")
		if !responseSent {
			ErrorResponse(w, r, http.StatusInternalServerError, l.InternalServerError)
			responseSent = true
		}
		return
	}

	finalSession.Values[a.UserIDSessionKey] = appUserID
	if dbUserForSession.DisplayName.Valid && dbUserForSession.DisplayName.String != "" {
		finalSession.Values[a.UserNameSessionKey] = dbUserForSession.DisplayName.String
	} else {
		finalSession.Values[a.UserNameSessionKey] = dbUserForSession.Username
	}

	// Double check linking flags are cleared to prevent orphaned state
	delete(finalSession.Values, a.LinkingUserIDSessionKey)
	delete(finalSession.Values, a.LinkingProviderNameSessionKey)

	// Save the session
	sessionSaveErr := finalSession.Save(r, w)
	if sessionSaveErr != nil {
		handleErr = fmt.Errorf("failed to save final application session: %w", sessionSaveErr)
		logger.Error().Err(handleErr).Msg("Critical: Failed to save the application session")
		if !responseSent {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save your session, please try again")
			responseSent = true
		}
		return
	}

	if handleErr == nil {
		logger.Info().Str(l.UserIDKey, appUserID).Msg("User session created/updated and saved successfully. Login/linking process complete.")
	} else {
		// This case should ideally be caught by the defer's commit check, but as a fallback, if handleErr is set (e.g. from a failed commit that didn't panic)
		// we log it here.
		logger.Error().Err(handleErr).Str(l.UserIDKey, appUserID).Msg("Login/linking process completed with an error during finalization (e.g. commit error).")
		return
	}

	// --- Step 7 ---
	// Redirect user to frontend

	if !strings.HasSuffix(redirectURL, "/") && !strings.Contains(redirectURL, "?") && !strings.HasSuffix(redirectURL, "dashboard") && !strings.HasSuffix(redirectURL, "settings") {
		// Avoid double slashes if already ends with settings or dashboard
		if redirectURL == ah.Cfg.FrontendBaseURL {
			redirectURL += "/"
		}
	}

	logger.Info().Str("redirect_url", redirectURL).Msg("Redirecting user to the frontend.")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Clears the user's session and logs them out
// Path: /auth/logout
// Method: GET or POST
func (ah *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(l.GetLoggerFromContext(ctx), authComponent, "HandleLogout")
	gothic.Logout(w, r)
	logger.Info().Msg("User logged out")
	return
}

// Retrieves the currently authenticated user's details
// Path: /auth/me
// Method: GET
func (ah *AuthHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		authComponent,
		"HandleGetCurrentUser",
	)

	session, err := gothic.Store.Get(r, a.SessionCookieName)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get session store")
		ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: session error")
		return
	}

	userIDVal, ok := session.Values[a.UserIDSessionKey]
	if !ok || userIDVal == nil {
		logger.Error().
			Str(l.UserIDKey, fmt.Sprintf("%v", userIDVal)).
			Msg("No user_id found in session or user_id is nil")
		ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: no user ID in session")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		logger.Warn().Str(l.UserIDKey, userID).Msg("user_id in session is not a string or is empty")
		ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: invalid user ID in session")
		return
	}

	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	dbUser, dbErr := db.GetUserByID(ctx, nil, userID)
	if dbErr != nil {
		if errors.Is(dbErr, db.ErrUserNotFound) {
			ErrorResponse(w, r, http.StatusUnauthorized, "Not authenticated: user not found")
		} else {
			logger.Error().Err(dbErr).Msg("Database error fetching user by ID")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve user details")
		}
		return
	}

	apiUser, convErr := modelConverters.DBUserToAPIUser(dbUser)
	if convErr != nil {
		logger.Error().Err(convErr).Str(l.UserIDKey, dbUser.UserID).Msg("Failed to convert database user to API user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process user details")
		return
	}

	meResponse := apiModels.AuthenticatedUserResponse{
		User: *apiUser,
	}

	Respond(w, r, http.StatusOK, meResponse, "Successfully retrieved user details")
}
