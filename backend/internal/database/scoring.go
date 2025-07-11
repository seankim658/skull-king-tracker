package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	dbModels "github.com/seankim658/skullking/internal/models/database"
	"github.com/seankim658/skullking/internal/rules"
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

// Handles submitting player scores, calculating results, updating totals, completing the round, and then
// creating the next round or completing the game
func SubmitScoresAndUpdateRound(
	ctx context.Context,
	tx *sql.Tx,
	gameID, roundID string,
	scores []dbModels.PlayerScoreData,
) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"SubmitScoresAndUpdateRound",
	).With().Str(l.RoundIDKey, roundID).Str(l.GameIDKey, gameID).Logger()

	// 1. Get current round details
	roundDetailsQuery := "SELECT round_number, dealer_game_player_id, is_tiebreaker_round FROM rounds WHERE round_id = $1;"
	logger.Debug().Str(l.QueryKey, roundDetailsQuery).Msg("Attempting to retrieve round details")

	var roundNumber int
	var dealerID string
	var isTiebreakerRound bool
	err := querier.QueryRowContext(ctx, roundDetailsQuery, roundID).Scan(&roundNumber, &dealerID, &isTiebreakerRound)
	if err != nil {
		return fmt.Errorf("failed to get round details: %w", err)
	}

	// 2. Get all the bids
	bidQuery := "SELECT game_player_id, bid_amount FROM player_round_scores WHERE round_id = $1;"
	rows, err := querier.QueryContext(ctx, bidQuery, roundID)
	if err != nil {
		return fmt.Errorf("failed to fetch bids for round: %w", err)
	}
	defer rows.Close()

	bids := make(map[string]int)
	for rows.Next() {
		var playerID string
		var bid sql.NullInt32
		if err := rows.Scan(&playerID, &bid); err != nil {
			return fmt.Errorf("failed to scan bid row: %w", err)
		}
		if !bid.Valid {
			return fmt.Errorf("player %s is missing a bid for this round", playerID)
		}
		bids[playerID] = int(bid.Int32)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating bid rows: %w", err)
	}

	// 3. Update scores for each player
	updatePlayerStmt, err := querier.PrepareContext(
		ctx,
		`
    UPDATE player_round_scores
    SET tricks_taken = $1, bonus_points_applied = $2, round_score = $3, updated_at = NOW()
    WHERE round_id = $4 AND game_player_id = $5;
    `,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare player score update statement: %w", err)
	}
	defer updatePlayerStmt.Close()

	updateTotalStmt, err := querier.PrepareContext(
		ctx,
		`UPDATE game_players SET final_score = final_score + $1 WHERE game_player_id = $2;`,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare player total score update statement: %w", err)
	}
	defer updateTotalStmt.Close()

	for _, scoreData := range scores {
		bidAmount, ok := bids[scoreData.GamePlayerID]
		if !ok {
			return fmt.Errorf("cannot score round, player %s has no bid record", scoreData.GamePlayerID)
		}

		effectiveRoundNumber := roundNumber
		if isTiebreakerRound {
			effectiveRoundNumber = 10
		}
		calculatedScore := rules.CalculatePlayerRoundScore(roundNumber, bidAmount, scoreData.TricksTaken)
		totalRoundScore := calculatedScore + scoreData.BonusPoints

		if _, err := updatePlayerStmt.ExecContext(
			ctx, scoreData.TricksTaken, scoreData.BonusPoints,
			totalRoundScore, roundID, scoreData.GamePlayerID,
		); err != nil {
			return fmt.Errorf("failed to update round score for player %s: %w", scoreData.GamePlayerID, err)
		}
		if _, err := updateTotalStmt.ExecContext(
			ctx, totalRoundScore, scoreData.GamePlayerID,
		); err != nil {
			return fmt.Errorf("failed to update total score for player %s: %w", scoreData.GamePlayerID, err)
		}
	}

	// 4. Mark current round as completed
	if _, err := querier.ExecContext(
		ctx,
		"UPDATE rounds SET status = 'completed', updated_at = NOW() WHERE round_id = $1;",
		roundID,
	); err != nil {
		return fmt.Errorf("failed to complete round: %w", err)
	}

	// 5. Transition to next state (next round or game over)
	if roundNumber >= 10 {
		tiedPlayerIDs, tieErr := CheckForTie(ctx, tx, gameID)
		if tieErr != nil {
			return fmt.Errorf("failed to check for ties: %w", tieErr)
		}

		if len(tiedPlayerIDs) > 1 {
			// Overtime triggered
			logger.Info().Msg("Tie detected, initiating overtime round")
			if err := UpdateGameStatus(ctx, tx, gameID, "overtime"); err != nil {
				return fmt.Errorf("failed to update game status to overtime: %w", err)
			}
			players, err := GetPlayersByGameID(ctx, tx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get players for dealer rotation: %w", err)
			}
			nextDealerID := getNextDealerID(players, dealerID)
			if _, err := CreateRound(ctx, tx, gameID, nextDealerID, roundNumber+1, true); err != nil {
				return fmt.Errorf("failed to create tiebreaker round %d: %w", roundNumber+1, err)
			}
		} else {
			logger.Info().Msg("No tie detected, marking game as finished")
			if err := UpdateGameStatus(ctx, tx, gameID, "completed"); err != nil {
				return fmt.Errorf("failed to mark game as completed: %w", err)
			}
			playerStats, err := GetGameSummaryStats(ctx, tx, gameID)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to get player stats for award calculation; awards will not be stored")
			} else {
				if len(playerStats) >= 4 {
					awards := rules.CalculateGameAwards(playerStats)
					if err := saveAwardsToDB(ctx, tx, gameID, awards); err != nil {
						logger.Error().Err(err).Msg("Failed to save game awards to database")
					}
				}
			}
		}
	} else {
		players, err := GetPlayersByGameID(ctx, tx, gameID)
		if err != nil {
			return fmt.Errorf("failed to get players for dealer rotation: %w", err)
		}
		nextDealerID := getNextDealerID(players, dealerID)
		nextRoundNumber := roundNumber + 1

		if _, err := CreateRound(ctx, tx, gameID, nextDealerID, nextRoundNumber, false); err != nil {
			return fmt.Errorf("failed to create round %d: %w", nextRoundNumber, err)
		}
		logger.Info().Int(l.RoundKey, nextRoundNumber).Msg("Next round created successfully")
	}

	logger.Info().Msg("Successfully submitted scores and transitioned game state")
	return nil
}

// Find the next dealer in the seating order
func getNextDealerID(players []dbModels.GamePlayerDetails, currentDealerID string) string {
	if len(players) == 0 {
		return ""
	}
	currentDealerIndex := -1
	for i, p := range players {
		if p.GamePlayerID == currentDealerID {
			currentDealerIndex = i
			break
		}
	}
	if currentDealerIndex == -1 {
		return players[0].GamePlayerID
	}
	return players[(currentDealerIndex+1)%len(players)].GamePlayerID
}

// Inserts calculated awards into the database
func saveAwardsToDB(ctx context.Context, tx *sql.Tx, gameID string, awards []apiModels.GameAward) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"saveAwardsToDB",
	).With().Str(l.GameIDKey, gameID).Logger()

	stmt, err := querier.PrepareContext(ctx, `
    INSERT INTO game_player_awards (game_id, game_player_id, user_id, award_type, award_value)
    SELECT $1, gp.game_player_id, $2, $3
    FROM game_players gp
    LEFT JOIN users u ON gp.user_id = u.user_id
    LEFT JOIN guest_players g ON gp.guest_player_id = g.guest_player_id
    WHERE gp.game_id = $1 AND COALESCE(u.display_name, u.username, g.display_name) = $4;
  `)
	if err != nil {
		return fmt.Errorf("failed to prepare award insert statement: %w", err)
	}
	defer stmt.Close()

	var awardsSaved int
	for _, award := range awards {
		awardTypeKey := strings.ToLower(strings.ReplaceAll(award.Title, "The ", ""))
		awardTypeKey = strings.ReplaceAll(awardTypeKey, " ", "-")

		if _, err := stmt.ExecContext(ctx, gameID, awardTypeKey, award.Value, award.PlayerName); err != nil {
			logger.Error().
				Err(err).
				Str(l.AwardTypeKey, award.Title).
				Str(l.GamePlayerNameKey, award.PlayerName).
				Msg("Failed to insert a game award")
		} else {
			awardsSaved++
		}
	}
	logger.Info().Int(l.CountKey, awardsSaved).Msg("Finished saving game awards")
	return nil
}
