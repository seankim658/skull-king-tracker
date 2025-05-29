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

	// Try to find existing guest player
	queryFind := "SELECT guest_player_id FROM guest_players WHERE display_name = $1;"
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
    game_id, session_id, created_by_user_id, current_scorekeeper_user_id, 
    status, starting_dealer_game_player_id, player_seating_order_randomized, 
    created_at, updated_at, completed_at
  FROM games
  WHERE game_id = $1;
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
