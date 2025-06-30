package database

import (
	"context"
	"database/sql"
	"fmt"

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
