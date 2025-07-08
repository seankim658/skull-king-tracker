package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const statsComponent = "database-stats"

// Retrieves the basic game statistics for a user
func GetUserBasicStats(ctx context.Context, tx *sql.Tx, userID string) (*dbModels.ProfileStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetUserBasicStats",
	).With().Str(l.UserIDKey, userID).Logger()

	stats := &dbModels.ProfileStats{}

	query := `
  WITH GamePlayerCounts AS (
    SELECT game_id, COUNT(*) as total_players
    FROM game_players
    GROUP BY game_id
  )
  SELECT
    COALESCE(COUNT(gp.game_id), 0),
    COALESCE(SUM(CASE WHEN gp.finishing_position = 1 THEN 1 ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN gp.finishing_position <= 3 AND gpc.total_players >= 4 THEN 1 ELSE 0 END), 0)
  FROM game_players gp
  JOIN games g ON gp.game_id = g.game_id
  LEFT JOIN GamePlayerCounts gpc ON g.game_id = gpc.game_id
  WHERE gp.user_id = $1 AND g.status = 'completed';
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get user profile stats")

	err := querier.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalGamesPlayed,
		&stats.TotalWins,
		&stats.Top3Finishes,
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user stats")
		return nil, fmt.Errorf("error getting stats for user %s: %w", userID, err)
	}

	logger.Info().Interface("base_profile_stats", stats).Msg("User basic stats retrieved successfully")
	return stats, nil
}

// Retrieves the basic site wide summary statistics
func GetSiteWideSummaryStats(ctx context.Context, tx *sql.Tx) (*dbModels.SiteWideSummaryStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetSiteWideSummaryStats",
	)

	stats := &dbModels.SiteWideSummaryStats{}
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

// Retrieves the game statistics for a user within a specific session
func GetUserSessionStats(ctx context.Context, tx *sql.Tx, userID, sessionID string) (*dbModels.ProfileStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetUserSessionStats",
	).With().Str(l.UserIDKey, userID).Str(l.SessionIDKey, sessionID).Logger()

	stats := &dbModels.ProfileStats{}

	query := `
  WITH GamePlayerCounts AS (
    SELECT game_id, COUNT(*) as total_players
    FROM game_players
    GROUP BY game_id
  )
  SELECT
    COALESCE(COUNT(gp.game_id), 0),
    COALESCE(SUM(CASE WHEN gp.finishing_position = 1 THEN 1 ELSE 0 END), 0),
    COALESCE(SUM(CASE WHEN gp.finishing_position <= 3 AND gpc.total_players >= 4 THEN 1 ELSE 0 END), 0)
  FROM game_players gp
  JOIN games g ON gp.game_id = g.game_id
  LEFT JOIN GamePlayerCounts gpc ON g.game_id = gpc.game_id
  WHERE gp.user_id = $1 AND g.session_id = $2 AND g.status = 'completed';
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get consolidated user session stats")

	err := querier.QueryRowContext(ctx, query, userID, sessionID).Scan(
		&stats.TotalGamesPlayed,
		&stats.TotalWins,
		&stats.Top3Finishes,
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user session stats")
		return nil, fmt.Errorf("error getting session stats for user %s in session %s: %w", userID, sessionID, err)
	}

	logger.Info().Interface("session_user_stats", stats).Msg("User session stats retrieved successfully")
	return stats, nil
}

// Retrieves and calculates all necessary stats for the end-of-game summary
func GetGameSummaryStats(ctx context.Context, tx *sql.Tx, gameID string) ([]dbModels.GameSummaryPlayerStats, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetGameSummaryStats",
	).With().Str(l.GameIDKey, gameID).Logger()

	query := `
  WITH PlayerGameStats AS (
    SELECT
      gp.game_player_id,
      gp.final_score,
      gp.finishing_position,
      COALESCE(u.display_name, u.username, g.display_name) as display_name,
      COALESCE(SUM(CASE WHEN prs.bid_amount = prs.tricks_taken THEN 1 ELSE 0 END), 0)::int as rounds_hit,
      COALESCE(SUM(CASE WHEN prs.bid_amount != prs.tricks_taken THEN 1 ELSE 0 END), 0)::int as rounds_missed,
      COALESCE(SUM(CASE WHEN prs.bid_amount = 0 AND prs.tricks_taken = 0 THEN 1 ELSE 0 END), 0)::int as zero_bids_hit,
      COALESCE(SUM(prs.bonus_points_applied), 0)::int as total_bonus,
      COALESCE(SUM(prs.tricks_taken), 0)::int as total_tricks_taken,
      COALESCE(STDDEV_SAMP(prs.bid_amount), 0.0) as bid_stddev,
      COALESCE(SUM(CASE WHEN prs.bid_amount = prs.tricks_taken AND prs.bid_amount > 0 THEN prs.round_score ELSE 0 END), 0)::int as points_from_correct_bids,
      COALESCE(SUM(CASE WHEN prs.bid_amount = prs.tricks_taken AND prs.bid_amount > 0 THEN prs.tricks_taken ELSE 0 END), 0)::int as tricks_from_correct_bids,
      COALESCE(AVG(prs.bid_amount), 0.0) as avg_bid
    FROM game_players gp
    LEFT JOIN player_round_scores prs ON gp.game_player_id = prs.game_player_id
    LEFT JOIN users u ON gp.user_id = u.user_id
    LEFT JOIN guest_players g ON gp.guest_player_id = g.guest_player_id
    WHERE gp.game_id = $1
    GROUP BY gp.game_player_id, u.display_name, u.username, g.display_name
  )
  SELECT * FROM PlayerGameStats ORDER BY finishing_position ASC NULLS LAST;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get game summary stats")

	rows, err := querier.QueryContext(ctx, query, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query for game summary stats")
		return nil, fmt.Errorf("error querying game summary stats: %w", err)
	}
	defer rows.Close()

	var stats []dbModels.GameSummaryPlayerStats
	for rows.Next() {
		var s dbModels.GameSummaryPlayerStats
		if err := rows.Scan(
			&s.GamePlayerID,
			&s.FinalScore,
			&s.FinishingPosition,
			&s.DisplayName,
			&s.RoundsHit,
			&s.RoundsMissed,
			&s.ZeroBidsHit,
			&s.TotalBonus,
			&s.TotalTricksTaken,
			&s.BidStdDev,
			&s.PointsFromCorrectBids,
			&s.TricksFromCorrectBids,
			&s.AvgBid,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan game summary stats row")
			return nil, fmt.Errorf("error scanning game summary stats row: %w", err)
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating over game summary stats row")
		return nil, fmt.Errorf("error iterating game summary stats row: %w", err)
	}

	logger.Info().Int(l.CountKey, len(stats)).Msg("Game summary stats retrieved successfully")
	return stats, nil
}
