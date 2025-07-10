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
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	query := `
  SELECT
    (SELECT COUNT(*) FROM users) AS total_players,
    (SELECT COUNT(*) FROM users WHERE created_at >= $1) AS new_users_this_month,
    (SELECT COUNT(*) FROM game_sessions WHERE created_at >= $1) AS sessions_this_month,
    (SELECT COUNT(*) FROM games WHERE created_at >= $1) AS games_this_month;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get site-wide summary stats")

	err := querier.QueryRowContext(ctx, query, oneMonthAgo).Scan(
		&stats.TotalPlayers,
		&stats.NewUsersThisMonth,
		&stats.SessionsThisMonth,
		&stats.GamesThisMonth,
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get site-wide summary stats")
		return nil, fmt.Errorf("error getting site-wide summary stats: %w", err)
	}

	logger.Info().Interface("site_summary_stats", stats).Msg("Site wide summary stats retrieved successfully")
	return stats, nil
}

// Retrieves the game statistics for a user within a specific session
func GetUserSessionStats(ctx context.Context, querier DBTX, userID, sessionID string) (*dbModels.ProfileStats, error) {
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
      COALESCE(AVG(prs.bid_amount), 0.0) as avg_bid,
      COALESCE(VAR_SAMP(prs.round_score), 0.0) as round_score_variance,
      COALESCE(SUM(CASE WHEN r.round_number >= 8 THEN prs.round_score ELSE 0 END), 0)::int as points_last_three_rounds,
      COALESCE(MAX(ABS(prs.bid_amount - prs.tricks_taken)), 0)::int as biggest_bust,
      COALESCE(SUM(CASE WHEN prs.bid_amount = 0 AND prs.tricks_taken != 0 THEN 1 ELSE 0 END), 0)::int as failed_zero_bids,
      (SELECT COUNT(*) FROM player_game_asterisks pga WHERE pga.game_player_id = gp.game_player_id)::int as total_asterisks
    FROM game_players gp
    LEFT JOIN player_round_scores prs ON gp.game_player_id = prs.game_player_id
    LEFT JOIN rounds r ON prs.round_id = r.round_id
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
			&s.RoundScoreVariance,
			&s.PointsLastThreeRounds,
			&s.BiggestBust,
			&s.FailedZeroBids,
			&s.TotalAsterisks,
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

// Retrieves a summary of all awards a user has won
func GetUserAwardsSummary(ctx context.Context, tx *sql.Tx, userID string) ([]dbModels.UserAwardStat, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		statsComponent,
		"GetUserAwardsSummary",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT award_type, COUNT(*) as award_count
  FROM game_player_awards
  WHERE user_id = $1
  GROUP BY award_type
  ORDER BY award_count DESC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to retrieve user awards summary")

	rows, err := querier.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query user awards summary")
		return nil, fmt.Errorf("error querying awards for user %s: %w", userID, err)
	}
	defer rows.Close()

	var awards []dbModels.UserAwardStat
	for rows.Next() {
		var award dbModels.UserAwardStat
		if err := rows.Scan(&award.AwardType, &award.AwardCount); err != nil {
			logger.Error().Err(err).Msg("Failed to scan award stat row")
			return nil, fmt.Errorf("error scanning award stat: %w", err)
		}
		awards = append(awards, award)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating award stat rows: %w", err)
	}

	logger.Info().Int(l.CountKey, len(awards)).Msg("User awards summary retrieved successfully")
	return awards, nil
}
