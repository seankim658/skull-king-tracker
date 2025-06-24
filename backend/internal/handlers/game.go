package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

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

	// Step 3: Add the creator as the first player
	_, opErr = db.AddPlayerToGame(ctx, tx, gameID, &userID, nil, 1)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to automatically add creator to game")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to add creator as player")
		return
	}

	// Step 4: Commit Transaction
	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for game creation: %w", err)
		logger.Error().Err(opErr).Msg("Transaction commit failed")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize game creation")
		return
	}
	logger.Debug().Msg("Transaction committed successfully for game creation")

	// Step 5: Fetch the Created Game
	createdGame, fetchErr := db.GetGameByID(ctx, nil, gameID)
	if fetchErr != nil {
		logger.Error().Err(fetchErr).Str(l.GameIDKey, gameID).Msg("Failed to fetch newly created game for response")
		Respond(w, r, http.StatusCreated, map[string]string{"game_id": gameID}, "Game created successfully, but full details could not be retrieved")
		return
	}

	apiGameResponse := apiModels.GameResponse{
		GameID:                   createdGame.GameID,
		Status:                   createdGame.Status,
		CreatedAt:                createdGame.CreatedAt,
		CreatedByUserID:          createdGame.CreatedByUserID,
		CurrentScoreKeeperUserID: &createdGame.CurrentScorekeeperUserID.String,
	}
	if createdGame.SessionID.Valid {
		apiGameResponse.SessionID = &createdGame.SessionID.String
	}
	Respond(w, r, http.StatusCreated, apiGameResponse, "Game created successfully")
}

// Handles adding a player (a registered user or guest) to a game
// Path: /games/{game_id}/players
// Method: POST
func (gh *GameHandler) HandleAddPlayerToGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleAddPlayerToGame",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
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

// Handles removing a player from a game
// Path: /games/{game_id}/players/{game_player_id}
// Method: DELETE
func (gh *GameHandler) HandleRemovePlayerFromGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleRemovePlayerFromGame",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	gamePlayerID, ok := PathVar(w, r, "game_player_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Str(l.GamePlayerIDKey, gamePlayerID).Logger()

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	_, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		return
	}

	if err := db.DeleteGamePlayer(ctx, nil, gamePlayerID); err != nil {
		if errors.Is(err, db.ErrGamePlayerNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Player not found in this game")
		} else {
			logger.Error().Err(err).Msg("Failed to delete player from game")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not remove player")
		}
		return
	}

	Respond(w, r, http.StatusNoContent, nil, "Player removed successfully")
}

// Handles starting a game by updating its status
// Path: /games/{game_id}/start
// Method: PUT
func (gh *GameHandler) HandleStartGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleStartGame",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	game, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		return
	}

	if game.Status != "pending" {
		ErrorResponse(w, r, http.StatusConflict, fmt.Sprintf("Cannot start game because its status is already '%s'", game.Status))
		return
	}

	// TODO : add logic to determine starting dealer and first round details, seating order, etc

	// TODO : should use a transaction
	if err := db.UpdateGameStatus(ctx, nil, gameID, "active"); err != nil {
		logger.Error().Err(err).Msg("Failed to update game status to active")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to start the game")
		return
	}

	Respond(w, r, http.StatusOK, nil, "Game started successfully")
}

// Handles fetching all players for a specific game
// Path: /games/{game_id}/players
// Method: GET
func (gh *GameHandler) HandleGetGamePlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleGetGamePlayers",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	dbPlayers, err := db.GetPlayersByGameID(ctx, nil, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get players for game")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve players for game")
		return
	}

	apiPlayers := make([]apiModels.GamePlayerResponse, len(dbPlayers))
	for i, p := range dbPlayers {
		apiPlayers[i] = apiModels.GamePlayerResponse{
			GamePlayerID: p.GamePlayerID,
			GameID:       p.GameID,
			DisplayName:  p.DisplayName,
			SeatingOrder: p.SeatingOrder,
			FinalScore:   p.FinalScore,
		}
		if p.UserID.Valid {
			apiPlayers[i].UserID = &p.UserID.String
		}
		if p.GuestPlayerID.Valid {
			apiPlayers[i].GuestPlayerID = &p.GuestPlayerID.String
		}
		if p.AvatarURL.Valid {
			apiPlayers[i].AvatarURL = &p.AvatarURL.String
		}
	}

	Respond(w, r, http.StatusOK, apiPlayers, "Players retrieved successfully")
}

// Handles fetching the details of a single game
// Path: /games/{game_id}/details
// Method: GET
func (gh *GameHandler) HandleGetGameDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleGetGameDetails",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	dbGame, err := db.GetGameByID(ctx, nil, gameID)
	if err != nil {
		if errors.Is(err, db.ErrGameNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Game not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve game details")
		}
		return
	}

	apiGameResponse := apiModels.GameResponse{
		GameID:                   dbGame.GameID,
		Status:                   dbGame.Status,
		CreatedAt:                dbGame.CreatedAt,
		CreatedByUserID:          dbGame.CreatedByUserID,
		CurrentScoreKeeperUserID: &dbGame.CurrentScorekeeperUserID.String,
	}
	if dbGame.SessionID.Valid {
		apiGameResponse.SessionID = &dbGame.SessionID.String
	}

	Respond(w, r, http.StatusOK, apiGameResponse, "Game details retrieved successfully")
}

// Handles updating game settings like scorekeeper and seating order
// Path: /games/{game_id}/settings
// Method: PUT
func (gh *GameHandler) HandleUpdateGameSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameHandlerComponent,
		"HandleUpdateGameSettings",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	_, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		return
	}

	var req apiModels.UpdateGameSettingsRequest
	if !ParseJSON(w, r, &req) {
		return
	}
	if !RequireFields(w, r, map[string]string{
		"scorekeeper_user_id": req.ScorekeeperUserID,
	}) {
		return
	}
	if len(req.OrderedPlayerIDs) == 0 {
		ErrorResponse(w, r, http.StatusBadRequest, "ordered_player_ids cannot be empty")
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to update game settings")
	if !txOk {
		return
	}

	var opErr error
	defer func() {
		if p := recover(); p != nil {
			opErr = fmt.Errorf("panic recovered: %v", p)
			logger.Error().Err(opErr).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic in HandleUpdateGameSettings")
		}
		if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction")
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Error().Err(rbErr).Msg("Transaction rollback failed")
			}
		}
	}()

	// --- Validate ---
	dbPlayers, err := db.GetPlayersByGameID(ctx, tx, gameID)
	if err != nil {
		opErr = err
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verify game players")
		return
	}

	// Validate scorekeeper
	isScorekeeperValid := false
	for _, p := range dbPlayers {
		if p.UserID.Valid && p.UserID.String == req.ScorekeeperUserID {
			isScorekeeperValid = true
			break
		}
	}
	if !isScorekeeperValid {
		opErr = errors.New("invalid scorekeeper_user_id")
		ErrorResponse(w, r, http.StatusBadRequest, "Selected scorekeeper is not a valid player in this game")
		return
	}

	// Validate seating order
	if len(req.OrderedPlayerIDs) != len(dbPlayers) {
		opErr = fmt.Errorf("player list length mismatch: expected %d, got %d", len(dbPlayers), len(req.OrderedPlayerIDs))
		ErrorResponse(w, r, http.StatusBadRequest, "Player list mismatch: ensure all players are included in the seating order")
		return
	}

	playerIDSet := make(map[string]bool)
	for _, p := range dbPlayers {
		playerIDSet[p.GamePlayerID] = true
	}
	for _, reqPlayerID := range req.OrderedPlayerIDs {
		if !playerIDSet[reqPlayerID] {
			opErr = fmt.Errorf("invalid player ID in seating order: %s", reqPlayerID)
			ErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid player ID %s found in seating order", reqPlayerID))
			return
		}
	}

	opErr = db.UpdateGameSettings(ctx, tx, gameID, req.ScorekeeperUserID, req.OrderedPlayerIDs)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save game settings")
		return
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for game settings update: %w", err)
		logger.Error().Err(opErr).Msg("Transaction commit failed")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize game settings update")
		return
	}

	Respond(w, r, http.StatusOK, nil, "Game settings updated successfully")
}
