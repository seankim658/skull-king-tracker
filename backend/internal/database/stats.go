package database

import (
	"context"
	"database/sql"
	"fmt"

	l "github.com/seankim658/skullking/internal/logger"
)

const profileComponent = "database-stats"

type ProfileStats struct {
	TotalGamesPlayed int
	TotalWins        int
}

// Retrieves the basic game statistics for a user
func GetUserBasicStats(ctx context.Context, tx *sql.Tx, userID string) (*ProfileStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		profileComponent,
		"GetUserBasicStats",
	).With().Str(l.UserIDKey, userID).Logger()

	profStats := &ProfileStats{}

	queryGamesPlayed := `
  SELECT COUNT(DISTINCT g.game_id)
  FROM games g
  JOIN game_players gp ON g.game_id = gp.game_id
  WHERE gp.user_id = $1 AND g.status = 'completed';
  `
	logger.Debug().Str(l.QueryKey, queryGamesPlayed).Msg("Attempting to get total games played")
	err := querier.QueryRowContext(ctx, queryGamesPlayed, userID).Scan(&profStats.TotalGamesPlayed)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total games played")
		return nil, fmt.Errorf("error getting total games played for user %s: %w", userID, err)
	}

	queryTotalWins := `
  SELECT COUNT(DISTINCT g.game_id)
  FROM games g
  JOIN game_players gp ON g.game_id = gp.game_id
  WHERE gp.user_id = $1 AND g.status = 'completed' AND gp.finishing_position = 1;
  `
	logger.Debug().Str(l.QueryKey, queryTotalWins).Msg("Attempting to get total wins")
	err = querier.QueryRowContext(ctx, queryTotalWins, userID).Scan(&profStats.TotalWins)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total wins")
		return nil, fmt.Errorf("error getting total wins for user %s: %w", userID, err)
	}

	logger.Info().Interface("base_profile_stats", profStats).Msg("User basic stats retrieved successfully")
	return profStats, nil
}
