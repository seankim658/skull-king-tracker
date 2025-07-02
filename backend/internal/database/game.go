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

const gameComponent = "database-game"

// Inserts a new game into the games table
func CreateGame(
	ctx context.Context,
	tx *sql.Tx,
	sessionID *string,
	createdByUserID,
	currentScorekeeperUserID,
	initialStatus string,
	playerSeatingOrderRandomized bool,
) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"CreateGame",
	).With().Str(l.UserIDKey, createdByUserID).Str(l.GameStatusKey, initialStatus).Logger()
	if sessionID != nil {
		logger.With().Str(l.SessionIDKey, *sessionID).Logger()
	}

	newGameID := uuid.NewString()
	currentTime := time.Now()

	query := `
  INSERT INTO games (
    game_id, session_id, created_by_user_id, current_scorekeeper_user_id, 
    status, player_seating_order_randomized, created_at, updated_at
  )
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING game_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create game")

	var sqlSessionID sql.NullString
	if sessionID != nil {
		sqlSessionID = NullString(*sessionID)
	}

	var returnedGameID string
	err := querier.QueryRowContext(ctx, query,
		newGameID,
		sqlSessionID,
		createdByUserID,
		NullString(currentScorekeeperUserID),
		initialStatus,
		playerSeatingOrderRandomized,
		currentTime,
		currentTime,
	).Scan(&returnedGameID)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create game")
		return "", fmt.Errorf("error creating game: %w", err)
	}

	logger.Info().Str(l.GameIDKey, returnedGameID).Msg("Game created successfully")
	return returnedGameID, nil
}

// Finds a guest by display name or creates a new one
func FindOrCreateGuestPlayer(ctx context.Context, tx *sql.Tx, displayName string) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"FindOrCreateGuestPlayer",
	).With().Str(l.GuestPlayerNameKey, displayName).Logger()

	queryFind := "SELECT guest_player_id FROM guest_players WHERE display_name = $1;"
	logger.Debug().Str(l.QueryKey, queryFind).Msg("Attempting to find existing guest player")
	var guestPlayerID string
	err := querier.QueryRowContext(ctx, queryFind, displayName).Scan(&guestPlayerID)
	if err == nil {
		logger.Debug().Str(l.GuestPlayerIDKey, guestPlayerID).Msg("Found existing guest player")
		return guestPlayerID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		logger.Error().Err(err).Msg("Error querying for existing guest player")
		return "", fmt.Errorf("error finding guest player %s: %w", displayName, err)
	}

	logger.Debug().Msg("Guest player not found, creating new one")
	newGuestPlayerID := uuid.NewString()
	currentTime := time.Now()
	queryCreate := `
  INSERT INTO guest_players (guest_player_id, display_name, created_at)
  VALUES ($1, $2, $3)
  RETURNING guest_player_id;
  `
	logger.Debug().Str(l.QueryKey, queryCreate).Msg("Attempting to create new guest player")
	err = querier.QueryRowContext(ctx, queryCreate, newGuestPlayerID, displayName, currentTime).Scan(&guestPlayerID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create new guest player")
		return "", fmt.Errorf("error creating guest player %s: %w", displayName, err)
	}

	logger.Info().Str(l.GuestPlayerIDKey, guestPlayerID).Msg("Guest player created successfully")
	return guestPlayerID, nil
}

// Adds a registered user to a game
func AddPlayerToGame(ctx context.Context, tx *sql.Tx, gameID string, userID, guestPlayerID *string, seatingOrder int) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"AddPlayerToGame",
	).With().Str(l.GameIDKey, gameID).Int(l.SeatingOrderKey, seatingOrder).Logger()
	if userID != nil {
		logger = logger.With().Str(l.UserIDKey, *userID).Logger()
	}
	if guestPlayerID != nil {
		logger = logger.With().Str(l.GuestPlayerIDKey, *guestPlayerID).Logger()
	}

	newGamePlayerID := uuid.NewString()

	query := `
  INSERT INTO game_players (
    game_player_id, game_id, user_id, guest_player_id, seating_order, final_score
  )
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING game_player_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to add player to game")

	var sqlUserID, sqlGuestPlayerID sql.NullString
	if userID != nil {
		sqlUserID = NullString(*userID)
	}
	if guestPlayerID != nil {
		sqlGuestPlayerID = NullString(*guestPlayerID)
	}

	var returnedGamePlayerID string
	err := querier.QueryRowContext(ctx, query,
		newGamePlayerID,
		gameID,
		sqlUserID,
		sqlGuestPlayerID,
		seatingOrder,
		0,
	).Scan(&returnedGamePlayerID)
	if err != nil {
		constraintMappings := map[string]error{
			"uq_game_user":  ErrPlayerAlreadyInGame,
			"uq_game_guest": ErrPlayerAlreadyInGame,
		}
		handled, appErr := HandlePgError(err, logger, constraintMappings)
		if handled {
			return "", appErr
		}
		logger.Error().Err(err).Msg("Failed to add player to game")
		return "", fmt.Errorf("error adding player to game %s: %w", gameID, err)
	}

	logger.Info().Str(l.GamePlayerIDKey, returnedGamePlayerID).Msg("Player added to game successfully")
	return returnedGamePlayerID, nil
}

// Deletes a player from a game using their game_player_id
func DeleteGamePlayer(ctx context.Context, tx *sql.Tx, gamePlayerID string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"DeleteGamePlayer",
	).With().Str(l.GamePlayerIDKey, gamePlayerID).Logger()

	query := `DELETE FROM game_players WHERE game_player_id = $1;`
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to delete game player")
	result, err := querier.ExecContext(ctx, query, gamePlayerID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete game player")
		return fmt.Errorf("error deleting game player %s: %w", gamePlayerID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrGamePlayerNotFound
	}

	logger.Info().Msg("Game player deleted successfully")
	return nil
}

// Retrieves a game by its ID
func GetGameByID(ctx context.Context, tx *sql.Tx, gameID string) (*dbModels.Game, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"GetGameByID",
	).With().Str(l.GameIDKey, gameID).Logger()

	query := `
  SELECT
    g.game_id, g.session_id, g.created_by_user_id, g.current_scorekeeper_user_id,
    g.status, g.starting_dealer_game_player_id, g.player_seating_order_randomized,
    g.created_at, g.updated_at, g.completed_at,
    gs.session_name,
    COALESCE(u.display_name, u.username) as scorekeeper_name
  FROM games g
  LEFT JOIN game_sessions gs ON g.session_id = gs.session_id
  LEFT JOIN users u ON g.current_scorekeeper_user_id = u.user_id
  WHERE g.game_id = $1;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get game by ID")

	game, err := scanGame(querier.QueryRowContext(ctx, query, gameID))
	if err != nil {
		if errors.Is(err, ErrGameNotFound) {
			logger.Warn().Msg("Game not found by ID")
		}
		return nil, err
	}
	logger.Info().Msg("Game retrieved successfully by ID")
	return game, nil
}

// Updates the status of a game
func UpdateGameStatus(ctx context.Context, tx *sql.Tx, gameID, status string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"UpdateGameStatus",
	).With().Str(l.GameIDKey, gameID).Str(l.GameStatusKey, status).Logger()

	query := `UPDATE games SET status = $1, updated_at = NOW() WHERE game_id = $2;`
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update game status")
	result, err := querier.ExecContext(ctx, query, status, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update game status")
		return fmt.Errorf("error updating game status for %s: %w", gameID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrGameNotFound
	}

	logger.Info().Msg("Game status updated successfully")
	return nil
}

// Retrieves all players for a given game
func GetPlayersByGameID(ctx context.Context, tx *sql.Tx, gameID string) ([]dbModels.GamePlayerDetails, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"GetPlayersByGameID",
	).With().Str(l.GameIDKey, gameID).Logger()

	query := `
  SELECT
    gp.game_player_id,
    gp.game_id,
    gp.user_id,
    gp.guest_player_id,
    COALESCE(u.display_name, u.username, g.display_name) AS display_name,
    u.avatar_url,
    gp.seating_order,
    gp.final_score
  FROM game_players gp
  LEFT JOIN users u ON gp.user_id = u.user_id
  LEFT JOIN guest_players g ON gp.guest_player_id = g.guest_player_id
  WHERE gp.game_id = $1
  ORDER BY gp.seating_order ASC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get players by game ID")

	rows, err := querier.QueryContext(ctx, query, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query game players")
		return nil, fmt.Errorf("error querying players for game %s: %w", gameID, err)
	}
	defer rows.Close()

	var players []dbModels.GamePlayerDetails
	for rows.Next() {
		var p dbModels.GamePlayerDetails
		if err := rows.Scan(
			&p.GamePlayerID,
			&p.GameID,
			&p.UserID,
			&p.GuestPlayerID,
			&p.DisplayName,
			&p.AvatarURL,
			&p.SeatingOrder,
			&p.FinalScore,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan game player row")
			return nil, fmt.Errorf("error scanning player data for game %s: %w", gameID, err)
		}
		players = append(players, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over game player rows for game %s: %w", gameID, err)
	}

	logger.Info().Int(l.CountKey, len(players)).Msg("Game players retrieved successfully")
	return players, nil
}

// Retrieves all games for a given session, including winner information
func GetGamesBySessionID(
	ctx context.Context,
	tx *sql.Tx,
	sessionID, viewerID string,
) ([]dbModels.GameWithWinner, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"GetGamesBySessionID",
	).With().Str(l.SessionIDKey, sessionID).Str(l.UserIDKey, viewerID).Logger()

	query := `
  SELECT
    g.game_id,
    g.status,
    g.created_at,
    g.completed_at,
    COALESCE(u.display_name, u.username, gp_winner.display_name) AS winning_player,
    (g.current_scorekeeper_user_id = $2) AS is_viewer_scorekeeper,
    COALESCE(u_sk.display_name, u_sk.username) as scorekeeper_name
  FROM games g
  LEFT JOIN game_players p ON g.game_id = p.game_id AND p.finishing_position = 1
  LEFT JOIN users u ON p.user_id = u.user_id
  LEFT JOIN guest_players gp_winner ON p.guest_player_id = gp_winner.guest_player_id
  LEFT JOIN users u_sk ON g.current_scorekeeper_user_id = u_sk.user_id
  WHERE g.session_id = $1
  ORDER BY g.created_at DESC;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to get games by session ID")

	rows, err := querier.QueryContext(ctx, query, sessionID, viewerID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query games by session ID")
		return nil, fmt.Errorf("error querying games for session %s: %w", sessionID, err)
	}
	defer rows.Close()

	var games []dbModels.GameWithWinner
	for rows.Next() {
		var game dbModels.GameWithWinner
		if err := rows.Scan(
			&game.GameID,
			&game.Status,
			&game.CreatedAt,
			&game.CompletedAt,
			&game.WinningPlayer,
			&game.IsViewerScorekeeper,
			&game.ScorekeeperName,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan game with winner row")
			return nil, fmt.Errorf("error scanning game data for session %s: %w", sessionID, err)
		}
		games = append(games, game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over games rows for session %s: %w", sessionID, err)
	}

	logger.Info().Int(l.CountKey, len(games)).Msg("Games for session retrieved successfully")
	return games, nil
}

// Updates the scorekeeper and seating order for a game
func UpdateGameSettings(
	ctx context.Context,
	querier DBTX,
	gameID, scorekeeperUserID,
	startingDealerGamePlayerID string,
	orderedPlayersIDs []string,
) error {
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"UpdateGameSettings",
	).With().
		Str(l.GameIDKey, gameID).
		Str(l.ScorekeeperIDKey, scorekeeperUserID).
		Str(l.StartingDealerPlayerIDKey, startingDealerGamePlayerID).
		Logger()

	updateScorekeeperQuery := `
  UPDATE games 
  SET 
    current_scorekeeper_user_id = $1, 
    starting_dealer_game_player_id = $2,
    updated_at = NOW()
  WHERE game_id = $3;
  `
	logger.Debug().Str(l.QueryKey, updateScorekeeperQuery).Msg("Attempting to update game scorekeeper")
	if _, err := querier.ExecContext(
		ctx,
		updateScorekeeperQuery,
		scorekeeperUserID,
		startingDealerGamePlayerID,
		gameID,
	); err != nil {
		logger.Error().Err(err).Msg("Failed to update scorekeeper")
		return fmt.Errorf("error updating scorekeeper for game %s: %w", gameID, err)
	}

	updateOrderQuery := `
  UPDATE game_players
  SET seating_order = $1, final_score = $2
  WHERE game_player_id = $3;
  `
	logger.Debug().Str(l.QueryKey, updateOrderQuery).Msg("Preparing to update seating order for players")
	stmt, err := querier.PrepareContext(ctx, updateOrderQuery)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to prepare statement for updating seating order")
		return fmt.Errorf("error preparing seating order update statement: %w", err)
	}
	defer stmt.Close()

	for i, playerID := range orderedPlayersIDs {
		seatingOrder := i + 1
		if _, err := stmt.ExecContext(ctx, seatingOrder, 0, playerID); err != nil {
			logger.Error().
				Err(err).
				Str(l.GamePlayerIDKey, playerID).
				Int(l.SeatingOrderKey, seatingOrder).
				Msg("Failed to update seating order")
			return fmt.Errorf("error updating seating order for player %s: %w", playerID, err)
		}
	}

	logger.Info().Int("players_updated", len(orderedPlayersIDs)).Msg("Game settings updated successfully")
	return nil
}

// Updates a game's status and sets the starting dealer
func UpdateGameStartDetails(
	ctx context.Context,
	tx *sql.Tx,
	gameID,
	startingDealerGamePlayerID,
	status string) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"UpdateGameStartDetails",
	).With().Str(l.GameIDKey, gameID).Str("starting_dealer_id", startingDealerGamePlayerID).Logger()

	query := `
  UPDATE games
  SET status = $1, starting_dealer_game_player_id = $2, updated_at = NOW()
  WHERE game_id = $3;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to update game start details")
	result, err := querier.ExecContext(ctx, query, status, startingDealerGamePlayerID, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update game start details")
		return fmt.Errorf("error updating game start details for %s: %w", gameID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrGameNotFound
	}

	logger.Info().Msg("Game start details updated successfully")
	return nil
}

func CreateRound(
	ctx context.Context,
	tx *sql.Tx,
	gameID, dealerGamePlayerID string,
	roundNumber int,
	isTiebreaker bool,
) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"CreateRound",
	).With().Str(l.GameIDKey, gameID).Int(l.RoundKey, roundNumber).Logger()

	newRoundID := uuid.NewString()
	initialStatus := "bidding"

	query := `
  INSERT INTO rounds (
    round_id, game_id, round_number, dealer_game_player_id, 
    status, is_tiebreaker_round, created_at, updated_at
  )
  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
  RETURNING round_id;
  `
	logger.Debug().Str(l.QueryKey, query).Msg("Attempting to create round")

	var returnedRoundID string
	err := querier.QueryRowContext(ctx, query,
		newRoundID,
		gameID,
		roundNumber,
		dealerGamePlayerID,
		initialStatus,
		isTiebreaker,
	).Scan(&returnedRoundID)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to create round")
		return "", fmt.Errorf("error creating round for game %s: %w", gameID, err)
	}

	logger.Info().Str(l.RoundIDKey, returnedRoundID).Msg("Round created successfully")
	return returnedRoundID, nil
}

// Fetches all necessary data to build the game scorecard
func GetScorecardState(ctx context.Context, tx *sql.Tx, gameID string) (*dbModels.FullScorecardData, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"GetScorecardStats",
	).With().Str(l.GameIDKey, gameID).Logger()

	scorecardData := &dbModels.FullScorecardData{}

	game, err := GetGameByID(ctx, tx, gameID)
	if err != nil {
		return nil, fmt.Errorf("error getting game for scorecard: %w", err)
	}
	scorecardData.Game = *game

	players, err := GetPlayersByGameID(ctx, tx, gameID)
	if err != nil {
		return nil, fmt.Errorf("error getting players for scorecard: %w", err)
	}
	scorecardData.Players = players

	queryRounds := `
  SELECT 
    round_id, game_id, round_number, dealer_game_player_id, 
    status, is_tiebreaker_round, created_at, updated_at 
  FROM rounds 
  WHERE game_id = $1 
  ORDER BY round_number ASC;
  `
	rows, err := querier.QueryContext(ctx, queryRounds, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query rounds for scorecard")
		return nil, fmt.Errorf("error getting rounds for scorecard: %w", err)
	}
	defer rows.Close()

	var rounds []dbModels.Round
	for rows.Next() {
		var r dbModels.Round
		if err := rows.Scan(
			&r.RoundID, &r.GameID, &r.RoundNumber, &r.DealerGamePlayerID,
			&r.Status, &r.IsTiebreakerRound, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan round row")
			return nil, fmt.Errorf("error scanning round data: %w", err)
		}
		rounds = append(rounds, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating round rows: %w", err)
	}
	scorecardData.Rounds = rounds

	queryScores := `
  SELECT 
    prs.player_round_score_id, prs.round_id, prs.game_player_id, 
    prs.bid_amount, prs.tricks_taken, prs.round_score, prs.bonus_points_applied
  FROM player_round_scores prs
  JOIN rounds r ON prs.round_id = r.round_id
  WHERE r.game_id = $1;
  `
	scoreRows, err := querier.QueryContext(ctx, queryScores, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query player scores for scorecard")
		return nil, fmt.Errorf("error getting player scores: %w", err)
	}
	defer scoreRows.Close()

	var scores []dbModels.PlayerRoundScoreDetails
	for scoreRows.Next() {
		var s dbModels.PlayerRoundScoreDetails
		if err := scoreRows.Scan(
			&s.PlayerRoundScoreID, &s.RoundID, &s.GamePlayerID, &s.BidAmount,
			&s.TricksTaken, &s.RoundScore, &s.BonusPointsApplied,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan player score row")
			return nil, fmt.Errorf("error scanning player score data: %w", err)
		}
		scores = append(scores, s)
	}
	if err = scoreRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score rows: %w", err)
	}
	scorecardData.Scores = scores

	logger.Info().Msg("Successfully fetched all scorecard data")
	return scorecardData, nil
}

// Checks if a registered user is a player in a game
func IsUserInGame(ctx context.Context, tx *sql.Tx, userID, gameID string) (bool, error) {
	querier := GetQuerier(tx)
	query := "SELECT EXISTS(SELECT 1 FROM game_players WHERE user_id = $1 AND game_id = $2);"
	var exists bool
	err := querier.QueryRowContext(ctx, query, userID, gameID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if user is in game: %w", err)
	}
	return exists, nil
}

type ActiveGameDetails struct {
	GameID          string
	SessionName     sql.NullString
	ScorekeeperName sql.NullString
	IsScorekeeper   bool
	CreatedAt       time.Time
	CurrentRound    sql.NullInt32
	PlayersJSON     []byte
}

// Retrieves all active games a user is participating in
func GetActiveGamesByUserID(ctx context.Context, tx *sql.Tx, userID string) ([]ActiveGameDetails, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"GetActiveGamesByUserID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  WITH GamePlayers AS (
    SELECT
      g.game_id,
      json_agg(json_build_object(
        'display_name', COALESCE(u.display_name, u.username, gp_guest.display_name),
        'avatar_url', u.avatar_url
      )) AS players
    FROM games g
    JOIN game_players gp ON g.game_id = gp.game_id
    LEFT JOIN users u ON gp.user_id = u.user_id
    LEFT JOIN guest_players gp_guest ON gp.guest_player_id = gp_guest.guest_player_id
    WHERE g.status = 'active'
    GROUP BY g.game_id
  )
  SELECT
    g.game_id,
    gs.session_name,
    COALESCE(u_sk.display_name, u_sk.username) AS scorekeeper_name,
    (g.current_scorekeeper_user_id = $1) AS is_scorekeeper,
    g.created_at,
    (SELECT MAX(r.round_number) FROM rounds r WHERE r.game_id = g.game_id) AS current_round,
    gp.players
  FROM games g
  JOIN GamePlayers gp ON g.game_id = gp.game_id
  LEFT JOIN users u_sk ON g.current_scorekeeper_user_id = u_sk.user_id
  LEFT JOIN game_sessions gs ON g.session_id = gs.session_id
  WHERE g.status = 'active'
  AND EXISTS (
    SELECT 1
    FROM game_players gp_check
    WHERE gp_check.game_id = g.game_id AND gp_check.user_id = $1
  )
  ORDER BY g.updated_at DESC;
  `

	rows, err := querier.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query for active games")
		return nil, err
	}
	defer rows.Close()

	var games []ActiveGameDetails
	for rows.Next() {
		var game ActiveGameDetails
		if err := rows.Scan(
			&game.GameID,
			&game.SessionName,
			&game.ScorekeeperName,
			&game.IsScorekeeper,
			&game.CreatedAt,
			&game.CurrentRound,
			&game.PlayersJSON,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan active game row")
			return nil, err
		}
		games = append(games, game)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}
