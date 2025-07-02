package database

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const scoringComponent = "database-scoring"

// Fetches the most recent round for a game
func GetCurrentRoundInfo(ctx context.Context, tx *sql.Tx, gameID string) (*dbModels.Round, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"GetCurrentRoundInfo",
	).With().Str(l.GameIDKey, gameID).Logger()

	query := `
  SELECT 
    round_id, game_id, round_number, dealer_game_player_id, 
    status, is_tiebreaker_round, created_at, updated_at 
  FROM rounds 
  WHERE game_id = $1 
  ORDER BY round_number DESC 
  LIMIT 1;
  `

	var r dbModels.Round
	err := querier.QueryRowContext(ctx, query, gameID).
		Scan(
			&r.RoundID, &r.GameID, &r.RoundNumber, &r.DealerGamePlayerID,
			&r.Status, &r.IsTiebreakerRound, &r.CreatedAt, &r.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRoundsFound
		}
		logger.Error().Err(err).Msg("Failed to get current round info")
		return nil, err
	}
	return &r, nil
}

// Inserts bid records and updates the round status
func SubmitBidsAndUpdateRoundStatus(
	ctx context.Context,
	tx *sql.Tx,
	roundID string,
	bids []dbModels.PlayerBidData,
) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"SubmitBidAndUpdateRoundStatus",
	).With().Str(l.RoundIDKey, roundID).Logger()

	stmt, err := querier.PrepareContext(ctx, `
  INSERT INTO player_round_scores (round_id, game_player_id, bid_amount)
  VALUES ($1, $2, $3);
  `)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to prepare bid insert statement")
		return err
	}
	defer stmt.Close()

	for _, bid := range bids {
		_, err := stmt.ExecContext(ctx, roundID, bid.GamePlayerID, bid.BidAmount)
		if err != nil {
			logger.Error().Err(err).Str(l.GamePlayerIDKey, bid.GamePlayerID).Msg("Failed to insert bid")
			return err
		}
	}

	updateQuery := "UPDATE rounds SET status = 'playing', updated_at = NOW() WHERE round_id = $1;"
	result, err := querier.ExecContext(ctx, updateQuery, roundID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update round status to 'playing'")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("round with ID %s not found for status update", roundID)
	}

	logger.Info().Int(l.BidAmountKey, len(bids)).Msg("Successfully submitted bids and updated round status")
	return nil
}

// Calculates the score for a single player for a round based on their bid and tricks taken
func calculatePlayerRoundScore(roundNumber, bid, tricks int) int {
	// Zero bid
	if bid == 0 && tricks == 0 {
		return roundNumber * 10
	}
	if bid == 0 && tricks != 0 {
		return roundNumber * -10
	}

	// Bid correct (non-zero)
	if bid > 0 && bid == tricks {
		return bid * 20
	}
	if bid > 0 && bid != tricks {
		return int(math.Abs(float64(bid-tricks))) * -10
	}
	return 0
}

// Submits player scores, calculates results, updates totals, and completes the round
func SubmitScoresAndUpdateRound(
	ctx context.Context,
	tx *sql.Tx,
	roundID string,
	scores []dbModels.PlayerScoreData,
) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"SubmitScoresAndUpdateRound",
	).With().Str(l.RoundIDKey, roundID).Logger()

	var roundNumber int
	err := querier.QueryRowContext(
		ctx,
		"SELECT round_number FROM rounds WHERE round_id = $1;",
		roundID,
	).Scan(&roundNumber)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get round number")
		return err
	}

	updatePlayerScoreStmt, err := querier.PrepareContext(ctx, `
    UPDATE player_round_scores
    SET tricks_taken = $1, bonus_points_applied = $2, round_score = $3, updated_at = NOW()
    WHERE round_id = $4 AND game_player_id = $5;
    `,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare player score update statement: %w", err)
	}
	defer updatePlayerScoreStmt.Close()

	updatePlayerTotalStmt, err := querier.PrepareContext(ctx, `
    UPDATE game_players
    SET final_score = final_score + $1
    WHERE game_player_id = $2;
    `,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare player total score update statement: %w", err)
	}
	defer updatePlayerTotalStmt.Close()

	for _, scoreData := range scores {
		var bid sql.NullInt32
		err := querier.QueryRowContext(
			ctx,
			"SELECT bid_amount FROM player_round_scores WHERE round_id = $1 AND game_player_id = $2;",
			roundID, scoreData.GamePlayerID,
		).Scan(&bid)
		if err != nil {
			return fmt.Errorf("failed to fetch bid for player %s: %w", scoreData.GamePlayerID, err)
		}
		if !bid.Valid {
			return fmt.Errorf("cannot score round, player %s has no bid", scoreData.GamePlayerID)
		}

		calculatedScore := calculatePlayerRoundScore(
			roundNumber,
			int(bid.Int32),
			scoreData.TricksTaken,
		)
		totalRoundScore := calculatedScore + scoreData.BonusPoints

		_, err = updatePlayerScoreStmt.ExecContext(
			ctx, scoreData.TricksTaken, scoreData.BonusPoints,
			totalRoundScore, roundID, scoreData.GamePlayerID,
		)
		if err != nil {
			return fmt.Errorf("failed to update round score for player %s: %w", scoreData.GamePlayerID, err)
		}

		_, err = updatePlayerTotalStmt.ExecContext(
			ctx, totalRoundScore,
			scoreData.GamePlayerID,
		)
		if err != nil {
			return fmt.Errorf("failed to update total score for player %s: %w", scoreData.GamePlayerID, err)
		}
	}

	_, err = querier.ExecContext(
		ctx,
		"UPDATE rounds SET status = 'completed', updated_at = NOW() WHERE round_id = $1;",
		roundID,
	)
	if err != nil {
		return fmt.Errorf("failed to complete round: %w", err)
	}

	logger.Info().Msg("Successfully submitted and calculated scores for the round")
	return nil
}
