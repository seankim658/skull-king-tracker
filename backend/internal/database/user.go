package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/markbates/goth"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const userComponent = "database-user"

// Inserts a new user into the database
func CreateUser(ctx context.Context, tx *sql.Tx, user *dbModels.User) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"CreateUser",
	).With().Str(l.UsernameKey, user.Username).Logger()

	newUserID := uuid.NewString()
	currentTime := time.Now()

	query := `
  INSERT INTO users (
    user_id, username, email, display_name, avatar_url, avatar_source,
    stats_privacy, created_at, updated_at, last_login_at
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
  RETURNING user_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create user")

	if user.StatsPrivacy == "" {
		user.StatsPrivacy = "public"
	}
	if strings.TrimSpace(user.Username) == "" {
		return "", errors.New("username cannot be empty")
	}

	var returnedUserID string
	err := querier.QueryRowContext(ctx, query,
		newUserID,
		user.Username,
		user.Email,
		user.DisplayName,
		user.AvatarURL,
		user.AvatarSource,
		user.StatsPrivacy,
		currentTime,
		currentTime,
		sql.NullTime{Time: currentTime, Valid: true},
	).Scan(&returnedUserID)

	if err != nil {
		constraintMappings := map[string]error{
			"uq_users_username": ErrUsernameTaken,
			"uq_users_email":    ErrEmailTaken,
		}
		handled, appErr := HandlePgError(err, logger, constraintMappings)
		if handled {
			return "", appErr
		}
		logger.Error().Err(err).Msg("Failed to create user (unhandled DB error)")
		return "", fmt.Errorf("error creating user: %w", err)
	}

	logger.Info().Str(l.UserIDKey, returnedUserID).Msg("User created successfully")
	return returnedUserID, nil
}

// Retrieves a user from the database by their `user_id`
func GetUserByID(ctx context.Context, tx *sql.Tx, userID string) (*dbModels.User, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"GetUserByID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT 
    user_id, username, email, display_name, avatar_url, avatar_source, 
    stats_privacy, ui_theme, color_theme, created_at, updated_at, last_login_at
  FROM users
  WHERE user_id = $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user by ID")

	user, err := scanUser(querier.QueryRowContext(ctx, query, userID))
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			logger.Error().Err(err).Msg("Failed to get user by ID (after scanUser)")
		} else {
			logger.Warn().Msg("User not found by ID (reported by scanUser)")
		}
		return nil, err
	}
	logger.Info().Msg("User retrieved successfully by ID")
	return user, nil
}

// Retrieves a user from the database by their email
func GetUserByEmail(ctx context.Context, tx *sql.Tx, email string) (*dbModels.User, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"GetUserByEmail",
	).With().Str(l.EmailKey, email).Logger()

	query := `
  SELECT
    user_id, username, email, display_name, avatar_url, avatar_source, 
    stats_privacy, ui_theme, color_theme, created_at, updated_at, last_login_at
  FROM users
  WHERE email = $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user by email")

	user, err := scanUser(querier.QueryRowContext(ctx, query, email))
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			logger.Error().Err(err).Msg("Failed to get user by email (after scanUser)")
		} else {
			logger.Warn().Msg("User not found by email (reported by scanUser)")
		}
		return nil, err
	}
	logger.Info().Msg("User retrieved successfully by email")
	return user, nil
}

// Links an OAuth provider's identity to a user
func CreateUserProviderIdentity(ctx context.Context, tx *sql.Tx, identity *dbModels.UserProviderIdentity) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"CreateUserProviderIdentity",
	).With().
		Str(l.UserIDKey, identity.UserID).
		Str(l.ProviderKey, identity.ProviderName).
		Str(l.ProviderUserIDKey, identity.ProviderUserID).
		Logger()

	newIdentityID := uuid.NewString()
	currentTime := time.Now()

	query := `
  INSERT INTO user_provider_identities (
    provider_identity_id, user_id, provider_name, provider_user_id,
    provider_email, provider_display_name, provider_avatar_url,
    created_at, updated_at
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
  RETURNING provider_identity_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create user provider identity")

	var returnedIdentityID string
	err := querier.QueryRowContext(ctx, query,
		newIdentityID,
		identity.UserID,
		identity.ProviderName,
		identity.ProviderUserID,
		identity.ProviderEmail,
		identity.ProviderDisplayName,
		identity.ProviderAvatarURL,
		currentTime,
		currentTime,
	).Scan(&returnedIdentityID)

	if err != nil {
		constraintMappings := map[string]error{
			"uq_user_providers_identities_provider":      ErrProviderIdentityConflict,
			"uq_user_providers_identities_user_provider": ErrProviderIdentityConflict,
		}

		handled, appErr := HandlePgError(err, logger, constraintMappings)
		if handled {
			return "", appErr
		}
		logger.Error().Err(err).Msg("Failed to create user provider identity (unhandled DB error)")
		return "", fmt.Errorf("error creating user provider identity: %w", err)
	}

	logger.Info().Str(l.ProviderIdentityIDKey, returnedIdentityID).Msg("User provider identity created successfully")
	return returnedIdentityID, nil
}

// Retrieves a provider identity by provider name and provider user ID
func GetUserProviderIdentity(ctx context.Context, tx *sql.Tx, providerName, providerUserID string) (*dbModels.UserProviderIdentity, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"GetUserProviderIdentity",
	).With().
		Str(l.ProviderKey, providerName).
		Str(l.ProviderUserIDKey, providerUserID).
		Logger()

	query := `
  SELECT
    provider_identity_id, user_id, provider_name, provider_user_id,
    provider_email, provider_display_name, provider_avatar_url,
    created_at, updated_at
  FROM user_provider_identities
  WHERE provider_name = $1 AND provider_user_id = $2;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user provider identity")

	identity := &dbModels.UserProviderIdentity{}
	err := querier.QueryRowContext(ctx, query, providerName, providerUserID).Scan(
		&identity.ProviderIdentityID,
		&identity.UserID,
		&identity.ProviderName,
		&identity.ProviderUserID,
		&identity.ProviderEmail,
		&identity.ProviderDisplayName,
		&identity.ProviderAvatarURL,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info().Msg("User provider identity not found")
			return nil, ErrUserProviderIdentityNotFound
		}
		logger.Error().Err(err).Msg("Failed to get user provider identity")
		return nil, fmt.Errorf("error getting provider identity for %s - %s: %w", providerName, providerUserID, err)
	}

	logger.Info().Msg("User provider identity retrieved successfully")
	return identity, nil
}

// Updates the `last_login_at` timestamp for a user
func UpdateUserLastLogin(ctx context.Context, tx *sql.Tx, userID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"UpdateUserLastLogin",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  UPDATE users
  SET last_login_at = $1, updated_at = $1
  WHERE user_id = $2;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update user last login time")

	currentTime := time.Now()
	result, err := querier.ExecContext(ctx, query, sql.NullTime{Time: currentTime, Valid: true}, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update user last login time")
		return fmt.Errorf("error updating last login for user %s: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after updating last login")
		return fmt.Errorf("error checking rows affected for user %s: %w", userID, err)
	}

	if rowsAffected == 0 {
		logger.Warn().Msg("No user found with ID to update last login time (or time was already current)")
		return fmt.Errorf("no user found with ID %s to update last login time", userID)
	}

	logger.Info().Msg("User last login time updated successfully")
	return nil
}

// Updates the UI and color theme for a given user
func UpdateUserThemeSettings(ctx context.Context, tx *sql.Tx, userID, uiTheme, colorTheme string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"UpdateUserThemeSettings",
	).With().Str(l.UserIDKey, userID).Str(l.UIThemeKey, uiTheme).Str(l.ColorThemeKey, colorTheme).Logger()

	query := `
  UPDATE users
  SET ui_theme = $1, color_theme = $2, updated_at = NOW()
  WHERE user_id = $3;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update user theme settings")

	result, err := querier.ExecContext(ctx, query, NullString(uiTheme), NullString(colorTheme), userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update user theme settings")
		return fmt.Errorf("error updating theme settings for user %s: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after updating theme settings")
		return fmt.Errorf("error checking rows affected for user %s theme update: %w", userID, err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Updates specific fields for a user, only non-nil fields will be changed
func UpdateUserProfile(ctx context.Context, tx *sql.Tx, userID string, updates map[string]any) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"UpdateUserProfile",
	).With().Str(l.UserIDKey, userID).Logger()

	if len(updates) == 0 {
		return nil
	}

	var querySetParts []string
	var queryArgs []any
	argCounter := 1

	for field, value := range updates {
		switch field {
		case "display_name":
			querySetParts = append(querySetParts, fmt.Sprintf("display_name = $%d", argCounter))
			queryArgs = append(queryArgs, value)
			argCounter++
		case "avatar_url":
			querySetParts = append(querySetParts, fmt.Sprintf("avatar_url = $%d", argCounter))
			queryArgs = append(queryArgs, value)
			argCounter++
		case "avatar_source":
			querySetParts = append(querySetParts, fmt.Sprintf("avatar_source = $%d", argCounter))
			queryArgs = append(queryArgs, value)
			argCounter++
		case "stats_privacy":
			strVal, ok := value.(string)
			if !ok || !IsValidStatsPrivacy(strVal) {
				return fmt.Errorf("invalid value for stats_privacy: %v", value)
			}
			querySetParts = append(querySetParts, fmt.Sprintf("stats_privacy = $%d", argCounter))
			queryArgs = append(queryArgs, strVal)
			argCounter++
		default:
			return fmt.Errorf("unsupported field for update: %s", field)
		}
	}

	if len(querySetParts) == 0 {
		logger.Warn().Msg("No valid fields provided for update after validation")
		return nil
	}

	querySetParts = append(querySetParts, fmt.Sprintf("updated_at = $%d", argCounter))
	queryArgs = append(queryArgs, time.Now())
	argCounter++

	queryArgs = append(queryArgs, userID)

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE users SET ")
	queryBuilder.WriteString(strings.Join(querySetParts, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE user_id = $%d;", argCounter))

	fullQuery := queryBuilder.String()
	logger.Debug().Str(l.QueryKey, fullQuery).Interface(l.ArgsKey, queryArgs).Msg("Attempting to update user profile")

	result, err := querier.ExecContext(ctx, fullQuery, queryArgs...)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update user profile")
		return fmt.Errorf("error updating user profile for user %s: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after updating user profile")
		return fmt.Errorf("error checking rows affected for user %s profile update: %w", userID, err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Retrieves all provider identities for a given user
func GetUserProviderIdentitiesByUserID(ctx context.Context, tx *sql.Tx, userID string) ([]dbModels.UserProviderIdentity, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"GetUserProviderIdentitiesByUserID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT
    provider_identity_id, user_id, provider_name, provider_user_id,
		provider_email, provider_display_name, provider_avatar_url,
		created_at, updated_at
  FROM user_provider_identities
  WHERE user_id = $1
  ORDER BY provider_name;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user provider identities by user ID")

	rows, err := querier.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query user provider identities")
		return nil, fmt.Errorf("error querying provider identities for user %s: %w", userID, err)
	}
	defer rows.Close()

	var identities []dbModels.UserProviderIdentity
	for rows.Next() {
		var identity dbModels.UserProviderIdentity
		if err := rows.Scan(
			&identity.ProviderIdentityID,
			&identity.UserID,
			&identity.ProviderName,
			&identity.ProviderUserID,
			&identity.ProviderEmail,
			&identity.ProviderDisplayName,
			&identity.ProviderAvatarURL,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan user provider identity row")
			return nil, fmt.Errorf("error scanning provider identity row for user %s: %w", userID, err)
		}
		identities = append(identities, identity)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over user provider identity row")
		return nil, fmt.Errorf("error iterating provider identity rows for user %s: %w", userID, err)
	}

	if len(identities) == 0 {
		// This shouldn't happen
		logger.Error().Msg("No provider identities found for user")
	} else {
		logger.Info().Int(l.CountKey, len(identities)).Msg("User provider identities retrived successfully")
	}
	return identities, nil
}

// Removes a specific provider identity for a user
func DeleteUserProviderIdentity(ctx context.Context, tx *sql.Tx, userID, providerName string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"DeleteUserProviderIdentity",
	).With().Str(l.UserIDKey, userID).Str(l.ProviderKey, providerName).Logger()

	var count int
	countQuery := "SELECT COUNT(*) FROM user_provider_identities WHERE user_id = $1"
	err := querier.QueryRowContext(ctx, countQuery, userID).Scan(&count)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count existing provider identities")
		return fmt.Errorf("error counting provider identities for user %s: %w", userID, err)
	}

	if count <= 1 {
		logger.Warn().Int(l.CountKey, count).Msg("Attempt to delete the last provider identity")
		return ErrDeleteLastProviderIdentity
	}

	deleteQuery := `
  DELETE FROM user_provider_identities
  WHERE user_id = $1 AND provider_name = $2;
  `
	logger.Debug().Str(l.QueryKey, deleteQuery).Msg("Attempting to delete user provider identity")

	result, err := querier.ExecContext(ctx, deleteQuery, userID, providerName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete user provider identity")
		return fmt.Errorf("error deleting provider identity (%s) for user %s: %w", providerName, userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after deleting provider identity")
		return fmt.Errorf("error checking rows affected for user %s provider %s deletion: %w", userID, providerName, err)
	}

	if rowsAffected == 0 {
		logger.Warn().Msg("No provider identity found to delete for the given user and provider")
		return ErrUserProviderIdentityNotFound
	}

	logger.Info().Msg("User provider identity deleted successfully")
	return nil
}

// Updates mutable details of an existing provider identity
func UpdateUserProviderIdentityDetails(ctx context.Context, tx *sql.Tx, providerIdentityID string, gothUser goth.User) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"UpdateUserProviderIdentityDetails",
	).With().Str(l.ProviderIdentityIDKey, providerIdentityID).Logger()

	query := `
  UPDATE user_provider_identities
  SET provider_email = $1,
    provider_display_name = $2,
    provider_avatar_url = $3,
    updated_at = NOW()
  WHERE provider_identity_id = $4;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update user provider identity details")

	result, err := querier.ExecContext(ctx, query,
		NullString(gothUser.Email),
		NullString(gothUser.Name),
		NullString(gothUser.AvatarURL),
		providerIdentityID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute update for user provider identity details")
		return fmt.Errorf("error updating provider identity details for ID %s: %w", providerIdentityID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after updating provider identity details")
		return fmt.Errorf("error checking rows affected for provider identity ID %s update: %w", providerIdentityID, err)
	}

	if rowsAffected == 0 {
		return ErrUserProviderIdentityNotFound
	}

	logger.Info().Msg("User provider identity details updated successfully")
	return nil
}

// Searches for users by username or display name
func SearchUsers(ctx context.Context, tx *sql.Tx, searchQuery string, limit int) ([]dbModels.UserSearchResult, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		userComponent,
		"SearchUsers",
	).With().Str(l.SearchQueryKey, searchQuery).Int(l.LimitKey, limit).Logger()

	if limit <= 0 {
		limit = 10
	}
	if searchQuery == "" {
		return []dbModels.UserSearchResult{}, nil
	}

	searchTerm := "%" + strings.ToLower(searchQuery) + "%"

	query := `
  SELECT user_id, username, display_name, avatar_url
  FROM users
  WHERE (LOWER(username) ILIKE $1 OR LOWER(display_name) ILIKE $1)
  ORDER BY
    CASE
      WHEN LOWER(username) = LOWER($2) THEN 1     -- Exact username match first
			WHEN LOWER(display_name) = LOWER($2) THEN 2 -- Exact display name match second
			WHEN LOWER(username) LIKE $3 THEN 3         -- Starts with username
			WHEN LOWER(display_name) LIKE $3 THEN 4     -- Starts with display name
			ELSE 5 -- Contains match
		END,
		username ASC
	LIMIT $4;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to search users")

	rows, err := querier.QueryContext(ctx,
		query,
		searchTerm,
		strings.ToLower(searchQuery),
		strings.ToLower(searchQuery)+"%",
		limit,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to search users")
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer rows.Close()

	var users []dbModels.UserSearchResult
	for rows.Next() {
		var u dbModels.UserSearchResult
		if err := rows.Scan(
			&u.UserID,
			&u.Username,
			&u.DisplayName,
			&u.AvatarURL,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan user search result row")
			return nil, fmt.Errorf("error scanning user search result: %w", err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over user search result rows")
		return nil, fmt.Errorf("error iterating user search results: %w", err)
	}

	logger.Info().Int(l.CountKey, len(users)).Msg("User search completed")
	return users, nil
}
