package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

// Creates a `sql.NullString` from a string
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// Validates the stats privacy is a valid value
func IsValidStatsPrivacy(value string) bool {
	switch value {
	case "private", "friends_only", "public":
		return true
	default:
		return false
	}
}

// Defines the interface for scanning a single row, satisfied by *sql.Row
type RowScanner interface {
	Scan(dest ...any) error
}

// Scans a user row
func scanUser(row RowScanner) (*dbModels.User, error) {
	user := &dbModels.User{}
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.StatsPrivacy,
		&user.UITheme,
		&user.ColorTheme,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user data: %w", err)
	}
	return user, nil
}

// Handles common error checks for PostgreSQL errors
func HandlePgError(dbErr error, logger zerolog.Logger, constraintMappings map[string]error) (handled bool, appError error) {
	var pgErr *pgconn.PgError
	if errors.As(dbErr, &pgErr) {
		logger.Warn().
			Str(l.PostgresErrorCodeKey, pgErr.Code).
			Str(l.PostgresConstraintKey, pgErr.ConstraintName).
			Str(l.PosgresErrorDetailKey, pgErr.Detail).
			Str(l.PostgresErrorMessageKey, pgErr.Message).
			Msg("PostgreSQL error occurred")

		if pgErr.Code == uniqueConstraintErrorCode {
			if specificAppError, ok := constraintMappings[pgErr.ConstraintName]; ok {
				return true, specificAppError
			}
			logger.Error().Msgf("Unhandled unique contraint violation: %s", pgErr.ConstraintName)
			return true, fmt.Errorf("data conflict: %s. %w", pgErr.Detail, dbErr)
		}
	}
	return false, dbErr
}
