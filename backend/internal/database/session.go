package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"CheckUserParticipatedInSession",
	).With().Str(l.UserIDKey, userID).Str(l.SessionIDKey, sessionID).Logger()

	query := `
  SELECT EXISTS (
    SELECT 1
    FROM game_players gp
    JOIN games g ON gp.game_id = g.game_id
    WHERE gp.user_id = $1 AND g.session_id = $2
  );
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to check if user participated in session")

	var participated bool
	err := querier.QueryRowContext(ctx, query, userID, sessionID).Scan(&participated)
	if err != nil {
		err = fmt.Errorf("error checking user participation for user %s in session %s: %w", userID, sessionID, err)
		logger.Error().Err(err).Msg("Failed to check if user participated in session")
		return false, err
	}
	return participated, nil
}

// Counts a user's completed sessions
func CountUserSessionHistory(ctx context.Context, tx *sql.Tx, userID string) (int64, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"CountUserSessionHistory",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT COUNT(DISTINCT gs.session_id)
  FROM game_sessions gs
  JOIN games g ON gs.session_id = g.session_id
  JOIN game_players gp ON g.game_id = gp.game_id
  WHERE gs.status = 'completed' AND gp.user_id = $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to count user session history")

	var totalCount int64
	err := querier.QueryRowContext(ctx, query, userID).Scan(&totalCount)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count user session history")
		return 0, fmt.Errorf("error counting history for user %s: %w", userID, err)
	}
	return totalCount, nil
}

// Retrieves a paginated list of a user's completed sessions with aggregated stats
func GetUserSessionHistory(
	ctx context.Context,
	tx *sql.Tx,
	userID, sortBy, sortOrder string,
	page, pageSize int,
) ([]dbModels.UserSessionHistoryRow, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"GetUserSessionHistory",
	).With().
		Str(l.UserIDKey, userID).
		Str(l.SortByKey, sortBy).
		Str(l.SortOrderKey, sortOrder).
		Int(l.PageKey, page).
		Int(l.PageSizeKey, pageSize).Logger()

	offset := (page - 1) * pageSize

	sortColumnMap := map[string]string{
		"date_completed":             "gs.completed_at",
		"number_of_games":            "number_of_games",
		"your_wins":                  "your_wins",
		"average_finishing_position": "avg_finishing_position",
	}

	orderByClause := "ORDER BY gs.completed_at DESC"
	if validColumn, ok := sortColumnMap[sortBy]; ok {
		orderDirection := "ASC"
		if strings.ToUpper(sortOrder) == "DESC" {
			orderDirection = "DESC"
		}
		orderByClause = fmt.Sprintf("ORDER BY %s %s", validColumn, orderDirection)
	}

	query := fmt.Sprintf(`
    SELECT
      gs.session_id,
      gs.session_name,
      gs.completed_at,
      COALESCE(u_creator.display_name, u_creator.username) as session_creator,
      COUNT(g.game_id) as number_of_games,
      SUM(CASE WHEN gp.finishing_position = 1 THEN 1 ELSE 0 END) as your_wins,
      COALESCE(AVG(gp.finishing_position), 0) as avg_finishing_position
    FROM game_sessions gs
    JOIN games g ON gs.session_id = g.session_id
    JOIN game_players gp ON g.game_id = gp.game_id
    LEFT JOIN users u_creator ON gs.created_by_user_id = u_creator.user_id
    WHERE gs.status = 'completed' AND gp.user_id = $1
    GROUP BY gs.session_id, u_creator.display_name, u_creator.username
    %s
    LIMIT $2 OFFSET $3;
  `, orderByClause)
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user session history")

	rows, err := querier.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query user session history")
		return nil, err
	}
	defer rows.Close()

	var history []dbModels.UserSessionHistoryRow
	for rows.Next() {
		var h dbModels.UserSessionHistoryRow
		var avgFinishingPos float64
		if err := rows.Scan(
			&h.SessionID, &h.SessionName, &h.DateCompleted, &h.SessionCreator,
			&h.NumberOfGames, &h.YourWins, &avgFinishingPos,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan user session history row")
			return nil, err
		}
		h.TotalFinishingPosition = int(avgFinishingPos * float64(h.NumberOfGames))
		history = append(history, h)
	}
	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over session history rows")
		return nil, fmt.Errorf("error iterating session history rows: %w", err)
	}

	return history, nil
}

// Fetches all data required for the session details page
func GetSessionDetails(
	ctx context.Context,
	tx *sql.Tx,
	sessionID, viewerID string,
) (*dbModels.SessionDetailData, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"GetSessionDetails",
	).With().Str(l.SessionIDKey, sessionID).Str(l.ViewerUserIDKey, viewerID).Logger()

	details := &dbModels.SessionDetailData{}

	// 1. Get core session details
	session, err := GetGameSessionByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	details.Session = *session

	// 2. Get all games for the session
	games, err := GetGamesBySessionID(ctx, querier, sessionID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get games for session details: %w", err)
	}
	details.Games = games

	// 3. Get the user-specific stats for this session
	userStats, err := GetUserSessionStats(ctx, querier, viewerID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats for session details: %w", err)
	}
	details.UserStats = *userStats

	logger.Info().Msg("Successfully fetched all components for session details")
	return details, nil
}

// Retrieves the IDs of active sessions that have not been updated since the given threshold
func GetStaleSessions(ctx context.Context, tx *sql.Tx, threshold time.Time) ([]string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionComponent,
		"GetStaleSessions",
	).With().Time("threshold", threshold).Logger()

	query := `
  SELECT session_id
  FROM game_sessions
  WHERE status = 'active' AND updated_at < $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get stale sessions")

	rows, err := querier.QueryContext(ctx, query, threshold)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query stale sessions")
		return nil, err
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			logger.Error().Err(err).Msg("Failed to scan stale session ID")
			return nil, err
		}
		sessionIDs = append(sessionIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessionIDs, nil
}
