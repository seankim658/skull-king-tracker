package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
)

const gameHandlerComponent = "handlers-game"

type GameHandler struct {
	Cfg *cf.Config
}

func NewGameHandler(cfg *cf.Config) *GameHandler {
	return &GameHandler{Cfg: cfg}
}

// Handles the creation of a new game
// Path: /games
// Method: POST
func (hg *GameHandler) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleCreateGame",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	var req apiModels.CreateGameRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to start transaction for creating game")
	if !txOk {
		return
	}

	var opErr error
	var gameID string
	var finalSessionID *string

	defer func() {
		if p := recover(); p != nil {
			logger.Error().Interface(l.PanicKey, p).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic recovered")
			_ = tx.Rollback()
			if opErr == nil && gameID == "" { // Check if an error was already handled or game creation failed before panic
				ErrorResponse(w, r, http.StatusInternalServerError, "Critical error processing game creation.")
			}
		} else if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction due to error in handler logic")
			_ = tx.Rollback()
		}
	}()

	// Step 1: Handle Session
	if req.SessionName != nil && *req.SessionName != "" {
		// 1.1: Session name was included, create new session
		createdSessionID, err := db.CreateGameSession(ctx, tx, *req.SessionName, userID)
		if err != nil {
			opErr = fmt.Errorf("failed to create new game session: %w", err)
			logger.Error().Err(opErr).Str(l.SessionNameKey, *req.SessionName).Msg("Error creating game session")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to create game session")
			return
		}
		finalSessionID = &createdSessionID
		logger.Info().Str(l.SessionIDKey, *finalSessionID).Str(l.SessionNameKey, *req.SessionName).Msg("New game session created")
	} else if req.SessionID != nil && *req.SessionID != "" {
		// 1.2: Session ID was included, create new session
		finalSessionID = req.SessionID
		logger.Info().Str(l.SessionIDKey, *finalSessionID).Msg("Using existing game session ID")
	}

	// Step 2: Create Game
	initialStatus := "pending"
	playerSeatingOrderRandomized := true

	gameID, opErr = db.CreateGame(ctx, tx, finalSessionID, userID, userID, initialStatus, playerSeatingOrderRandomized)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to create game in database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to create gaem")
		return
	}
	logger.Info().Str(l.GameIDKey, gameID).Msg("Game created in database")

	// Step 3: Commit Transaction
	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for game creation: %w", err)
		logger.Error().Err(opErr).Msg("Transaction commit failed")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize game creation")
		return
	}
	logger.Debug().Msg("Transaction committed successfully for game creation")

	// Step 4: Fetch the Created Game
	createdGame, fetchErr := db.GetGameByID(ctx, nil, gameID)
	if fetchErr != nil {
		logger.Error().Err(fetchErr).Str(l.GameIDKey, gameID).Msg("Failed to fetch newly created game for response")
		Respond(w, r, http.StatusCreated, map[string]string{"game_id": gameID}, "Game created successfully, but full details could not be retrieved")
		return
	}

	apiGameResponse := apiModels.GameResponse{
		GameID:          createdGame.GameID,
		Status:          createdGame.Status,
		CreatedAt:       createdGame.CreatedAt,
		CreatedByUserID: createdGame.CreatedByUserID,
	}
	if createdGame.SessionID.Valid {
		apiGameResponse.SessionID = &createdGame.SessionID.String
	}
	Respond(w, r, http.StatusCreated, apiGameResponse, "Game created successfully")
}

// Handles adding a player (a registered user or guest) to an existing game
// Path: /games/{game_id}/players
// Method: POST
func (gh *GameHandler) HandleAddPlayerToGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleAddPlayerToGame",
	)

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok || gameID == "" {
		ErrorResponse(w, r, http.StatusBadRequest, "Game ID is required")
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	// Step 1: Authentication and Authorization Checks
	authenticatedUserID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}
	logger = logger.With().Str(l.UserIDKey, authenticatedUserID).Logger()

	_, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, authenticatedUserID, logger)
	if !authorized {
		return
	}

	var req apiModels.AddPlayerToGameRequest
	if !ParseJSON(w, r, &req) {
		return
	}
	if (req.UserID == nil || *req.UserID == "") && (req.GuestName == nil || *req.GuestName == "") {
		ErrorResponse(w, r, http.StatusBadRequest, "Either user_id or guest_name must be provided")
		return
	}
	if req.UserID != nil && *req.UserID != "" && req.GuestName != nil && *req.GuestName != "" {
		ErrorResponse(w, r, http.StatusBadRequest, "Provide either user_id or guest_name, not both")
		return
	}
	if req.SeatingOrder <= 0 {
		ErrorResponse(w, r, http.StatusBadRequest, "Seating order must be a positive integer")
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to start transaction for adding player")
	if !txOk {
		return
	}

	var opErr error
	var gamePlayerID string
	var finalGuestPlayerID *string
	var playerDisplayName string

	defer func() {
		if p := recover(); p != nil {
			logger.Error().Interface(l.PanicKey, p).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic recovered")
			_ = tx.Rollback()
			if opErr == nil && gamePlayerID == "" {
				ErrorResponse(w, r, http.StatusInternalServerError, "Critical error processing player addition")
			}
		} else if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction due to error in handler logic")
			_ = tx.Rollback()
		}
	}()

	// 1: Handle Guest Player (if applicable)
	if req.GuestName != nil && *req.GuestName != "" {
		createdGuestID, err := db.FindOrCreateGuestPlayer(ctx, tx, *req.GuestName)
		if err != nil {
			opErr = fmt.Errorf("failed to find or create guest player: %w", err)
			logger.Error().Err(opErr).Str(l.GuestPlayerNameKey, *req.GuestName).Msg("Error with guest player")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process guest player")
			return
		}
		finalGuestPlayerID = &createdGuestID
		playerDisplayName = *req.GuestName
		logger.Info().Str(l.GuestPlayerIDKey, *finalGuestPlayerID).Msg("Guest player processed")
	}

	// 2: Add Player to Game
	gamePlayerID, opErr = db.AddPlayerToGame(ctx, tx, gameID, req.UserID, finalGuestPlayerID, req.SeatingOrder)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to add player to game in database")
		if errors.Is(opErr, db.ErrPlayerAlreadyInGame) {
			ErrorResponse(w, r, http.StatusConflict, "This player is already in the game")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to add player to game")
		}
		return
	}
	logger.Info().Str(l.GamePlayerIDKey, gamePlayerID).Msg("Player added to game in game_players table")

	// 3: Determine Display Name (if a registered user)
	if req.UserID != nil && *req.UserID != "" {
		dbUser, userErr := db.GetUserByID(ctx, tx, *req.UserID)
		if userErr != nil {
			opErr = fmt.Errorf("failed to fetch user details for display name: %w", userErr)
			logger.Error().Err(opErr).Str(l.UserIDKey, *req.UserID).Msg("Could not fetch user for display name")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve player details")
			return
		}
		if dbUser.DisplayName.Valid && dbUser.DisplayName.String != "" {
			playerDisplayName = dbUser.DisplayName.String
		} else {
			playerDisplayName = dbUser.Username
		}
	}

	// 4: Commit Transaction
	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for adding player: %w", err)
		logger.Error().Err(opErr).Msg("Transaction commit failed")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize adding player")
		return
	}
	logger.Debug().Msg("Transaction committed successfully for adding player")

	// 5: Send response
	apiPlayerResponse := apiModels.GamePlayerResponse{
		GamePlayerID: gamePlayerID,
		GameID:       gameID,
		DisplayName:  playerDisplayName,
		SeatingOrder: req.SeatingOrder,
		FinalScore:   0,
	}
	if req.UserID != nil {
		apiPlayerResponse.UserID = req.UserID
	}
	if finalGuestPlayerID != nil {
		apiPlayerResponse.GuestPlayerID = finalGuestPlayerID
	}

	Respond(w, r, http.StatusCreated, apiPlayerResponse, "Player added to game successfully")
}
