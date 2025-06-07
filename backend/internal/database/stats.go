package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	l "github.com/seankim658/skullking/internal/logger"
)

const statsComponent = "database-stats"

type ProfileStats struct {
	TotalGamesPlayed int
	TotalWins        int
}

// Retrieves the basic game statistics for a user
func GetUserBasicStats(ctx context.Context, tx *sql.Tx, userID string) (*ProfileStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
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

type SiteWideSummaryStats struct {
	TotalPlayers      int
	SessionsThisMonth int
	GamesThisMonth    int
	NewUsersThisMonth int
}

// Retrieves the basic site wide summary statistics
func GetSiteWideSummaryStats(ctx context.Context, tx *sql.Tx) (*SiteWideSummaryStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetSiteWideSummaryStats",
	)

	stats := &SiteWideSummaryStats{}
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	queryTotalPlayers := "SELECT COUNT(*) FROM users;"
	logger.Debug().Str(l.QueryKey, queryTotalPlayers).Msg("Attempting to get total players")
	if err := querier.QueryRowContext(ctx, queryTotalPlayers).Scan(&stats.TotalPlayers); err != nil {
		logger.Error().Err(err).Msg("Failed to get total players")
		return nil, fmt.Errorf("error getting total players: %w", err)
	}

	querySessionsLastMonth := `
  SELECT COUNT(*) FROM game_sessions
  WHERE created_at >= $1 AND created_at <= $2;
  `
	logger.Debug().Str(l.QueryKey, querySessionsLastMonth).Msg("Attempting to get sessions this month")
	if err := querier.QueryRowContext(
		ctx,
		querySessionsLastMonth,
		oneMonthAgo,
		now,
	).Scan(&stats.SessionsThisMonth); err != nil {
		logger.Error().Err(err).Msg("Failed to get sessions this month")
		return nil, fmt.Errorf("error getting sessions this mont: %w", err)
	}

	queryGamesLastMonth := `
  SELECT COUNT(*) FROM games
  WHERE created_at >= $1 AND created_at <= $2;
  `
	logger.Debug().Str(l.QueryKey, queryGamesLastMonth).Msg("Attempting to get games this month")
	if err := querier.QueryRowContext(
		ctx,
		queryGamesLastMonth,
		oneMonthAgo,
		now,
	).Scan(&stats.GamesThisMonth); err != nil {
		logger.Error().Err(err).Msg("Failed to get games this month")
		return nil, fmt.Errorf("error getting games this month: %w", err)
	}

	queryNewUsersThisMonth := `
  SELECT COUNT(*) FROM users
  WHERE created_at >= $1 AND created_at <= $2;
  `
	logger.Debug().Str(l.QueryKey, queryNewUsersThisMonth).Msg("Attempting to get new users this month")
	if err := querier.QueryRowContext(
		ctx,
		queryNewUsersThisMonth,
		oneMonthAgo,
		now,
	).Scan(&stats.NewUsersThisMonth); err != nil {
		logger.Error().Err(err).Msg("Failed to get new users this month")
		return nil, fmt.Errorf("error getting new users this month: %w", err)
	}

	logger.Info().Interface("site_summary_stats", stats).Msg("Site wide summary stats retrieved successfully")
	return stats, nil
}
