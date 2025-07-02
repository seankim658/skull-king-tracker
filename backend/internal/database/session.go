package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const sessionComponent = "database-session"

// Inserts a new game session into the game sessions table
func CreateGameSession(ctx context.Context, tx *sql.Tx, sessionName, createdByUserID string) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"CreateGameSession",
	).With().Str(l.SessionNameKey, sessionName).Str(l.UserIDKey, createdByUserID).Logger()

	newSessionID := uuid.NewString()
	currentTime := time.Now()
	initialStatus := "active"

	query := `
  INSERT INTO game_sessions (
    session_id, session_name, created_by_user_id, status, 
    created_at, updated_at, completed_at
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  RETURNING session_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create game session")

	var returnedSessionID string
	err := querier.QueryRowContext(ctx, query,
		newSessionID,
		NullString(sessionName),
		NullString(createdByUserID),
		initialStatus,
		currentTime,
		currentTime,
		sql.NullString{},
	).Scan(&returnedSessionID)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create game session")
		return "", fmt.Errorf("error creating game session: %w", err)
	}

	logger.Info().Str(l.SessionIDKey, returnedSessionID).Msg("Game session created successfully")
	return returnedSessionID, nil
}

// Retrieves all active or pending sessions for a given user where the user participated in at least one game (checks if any games within those sessions is currently 'pending' or 'active').
func GetActiveSessionsByUserID(
	ctx context.Context,
	tx *sql.Tx,
	userID string,
) ([]dbModels.GameSessionWithActivity, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"GetActiveSessionsByUserID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT
    gs.session_id,
    gs.session_name,
    gs.created_by_user_id,
    gs.status,
    gs.created_at,
    gs.updated_at,
    gs.completed_at,
    EXISTS (
      SELECT 1 FROM games g_check
      WHERE g_check.session_id = gs.session_id AND g_check.status = 'active'
    ) as has_active_game,
    EXISTS (
      SELECT 1 FROM games g_check_pending
      WHERE g_check_pending.session_id = gs.session_id AND g_check_pending.status = 'pending'
    ) as has_pending_game,
    COALESCE(u.display_name, u.username) as creator_name
  FROM game_sessions gs
  LEFT JOIN users u ON gs.created_by_user_id = u.user_id
  WHERE gs.status = 'active'
  AND EXISTS (
    SELECT 1
    FROM games g_user
    JOIN game_players gp ON g_user.game_id = gp.game_id
    WHERE g_user.session_id = gs.session_id
    AND gp.user_id = $1
  )
  ORDER BY gs.updated_at DESC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get active sessions for user")

	rows, err := querier.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query active sessions")
		return nil, fmt.Errorf("error querying active sessions for user %s: %w", userID, err)
	}
	defer rows.Close()

	var sessions []dbModels.GameSessionWithActivity
	for rows.Next() {
		var s dbModels.GameSessionWithActivity
		if err := rows.Scan(
			&s.SessionID,
			&s.SessionName,
			&s.CreatedByUserID,
			&s.Status,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.CompletedAt,
			&s.HasActiveGame,
			&s.HasPendingGame,
			&s.CreatorName,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan active session row")
			return nil, fmt.Errorf("error scanning/pending active session row for user %s: %w", userID, err)
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over active session rows")
		return nil, fmt.Errorf("error iterating active session rows for user %s: %w", userID, err)
	}

	logger.Info().Int(l.CountKey, len(sessions)).Msg("Active sessions retrieved successfully")
	return sessions, nil
}

// Updates the status and completed_at timestamp of a game session
func UpdateSessionStatus(
	ctx context.Context,
	tx *sql.Tx,
	sessionID, status string,
	completedAt sql.NullTime,
) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"UpdateSessionStatus",
	).With().Str(l.SessionIDKey, sessionID).Str(l.StatusKey, status).Logger()

	query := `
  UPDATE game_sessions
  SET status = $1, completed_at = $2, updated_at = NOW()
  WHERE session_id = $3;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update session status")

	result, err := querier.ExecContext(ctx, query, status, completedAt, sessionID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to udpate session status")
		return fmt.Errorf("error updating session status for session %s: %w", sessionID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get rows affected after updating session status")
		return fmt.Errorf("error checking rows affected for session %s status update: %w", sessionID, err)
	}

	if rowsAffected == 0 {
		logger.Warn().Msg("No session found with ID to update status (or status was already current)")
		return ErrSessionNotFound
	}

	logger.Info().Msg("Session status updated successfully")
	return nil
}

// Retrieves a game session by its ID
func GetGameSessionByID(ctx context.Context, tx *sql.Tx, sessionID string) (*dbModels.GameSession, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"GetGameSessionByID",
	).With().Str(l.SessionIDKey, sessionID).Logger()

	query := `
  SELECT session_id, session_name, created_by_user_id, status, created_at, updated_at, completed_at
  FROM game_sessions
  WHERE session_id = $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get game session by ID")

	session, err := scanGameSession(querier.QueryRowContext(ctx, query, sessionID))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			logger.Warn().Msg("Game session not found by ID")
		} else {
			logger.Error().Err(err).Msg("Failed to get game session by ID (after scan)")
		}
		return nil, err
	}
	logger.Info().Msg("Game session retrieved successfully by ID")
	return session, nil
}

// Checks if a user has participated in any game within a given session
func CheckUserParticipatedInSession(
	ctx context.Context, tx *sql.Tx, userID, sessionID string,
) (bool, error) {
	querier := GetQuerier(tx)
	query := `
  SELECT EXISTS (
    SELECT 1
    FROM game_players gp
    JOIN games g ON gp.game_id = g.game_id
    WHERE gp.user_id = $1 AND g.session_id = $2
  );
  `

	var participated bool
	err := querier.QueryRowContext(ctx, query, userID, sessionID).Scan(&participated)
	if err != nil {
		return false,
			fmt.Errorf(
				"error checking user participation for user %s in session %s: %w", userID, sessionID, err,
			)
	}
	return participated, nil
}
