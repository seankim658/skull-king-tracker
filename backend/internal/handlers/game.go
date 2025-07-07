package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
)

const gameComponent = "handlers-game"

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
		gameComponent,
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	var finalSessionID *string

	// Step 1: Handle Session
	if req.SessionName != nil && *req.SessionName != "" {
		// 1.1: Session name was included, create new session
		createdSessionID, err := db.CreateGameSession(ctx, tx, *req.SessionName, userID)
		if err != nil {
			opErr = fmt.Errorf("failed to create new game session: %w", err)
			logger.Error().
				Err(opErr).
				Str(l.SessionNameKey, *req.SessionName).
				Msg("Error creating game session")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to create game session")
			responseSent = true
			return
		}
		finalSessionID = &createdSessionID
		logger.Info().
			Str(l.SessionIDKey, *finalSessionID).
			Str(l.SessionNameKey, *req.SessionName).
			Msg("New game session created")
	} else if req.SessionID != nil && *req.SessionID != "" {
		// 1.2: Session ID was included, create new session
		finalSessionID = req.SessionID
		logger.Info().Str(l.SessionIDKey, *finalSessionID).Msg("Using existing game session ID")
	}

	// Step 2: Create Game
	initialStatus := "pending"
	playerSeatingOrderRandomized := true

	gameID, err := db.CreateGame(
		ctx, tx, finalSessionID,
		userID, userID, initialStatus,
		playerSeatingOrderRandomized,
	)
	if err != nil {
		opErr = fmt.Errorf("failed to create game in database: %w", err)
		logger.Error().Err(opErr).Msg("Failed to create game in database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to create gaem")
		responseSent = true
		return
	}
	logger.Info().Str(l.GameIDKey, gameID).Msg("Game created in database")

	// Step 3: Add the creator as the first player
	_, err = db.AddPlayerToGame(ctx, tx, gameID, &userID, nil, 1)
	if err != nil {
		opErr = fmt.Errorf("failed to automatically add creator to game: %w", err)
		logger.Error().Err(opErr).Msg("Failed to automatically add creator to game")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to add creator as player")
		responseSent = true
		return
	}

	if responseSent {
		return
	}

	// Step 5: Fetch the Created Game
	createdGame, fetchErr := db.GetGameByID(ctx, tx, gameID)
	if fetchErr != nil {
		opErr = fetchErr
		logger.Error().
			Err(fetchErr).
			Str(l.GameIDKey, gameID).
			Msg("Failed to fetch newly created game for response")
		Respond(
			w, r, http.StatusCreated,
			map[string]string{"game_id": gameID},
			"Game created successfully, but full details could not be retrieved",
		)
		responseSent = true
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
	responseSent = true
}

// Handles adding a player (a registered user or guest) to a game
// Path: /games/{game_id}/players
// Method: POST
func (gh *GameHandler) HandleAddPlayerToGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	var gamePlayerID string
	var finalGuestPlayerID *string
	var playerDisplayName string

	// 1: Handle Guest Player (if applicable)
	if req.GuestName != nil && *req.GuestName != "" {
		createdGuestID, err := db.FindOrCreateGuestPlayer(ctx, tx, *req.GuestName)
		if err != nil {
			opErr = fmt.Errorf("failed to find or create guest player: %w", err)
			logger.Error().Err(opErr).Str(l.GuestPlayerNameKey, *req.GuestName).Msg("Error with guest player")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to process guest player")
			responseSent = true
			return
		}
		finalGuestPlayerID = &createdGuestID
		playerDisplayName = *req.GuestName
		logger.Info().Str(l.GuestPlayerIDKey, *finalGuestPlayerID).Msg("Guest player processed")
	}

	// 2: Add Player to Game
	gamePlayerID, err := db.AddPlayerToGame(ctx, tx, gameID, req.UserID, finalGuestPlayerID, req.SeatingOrder)
	if err != nil {
		opErr = fmt.Errorf("failed to add player to game in database: %w", err)
		logger.Error().Err(opErr).Msg("Failed to add player to game in database")
		if errors.Is(opErr, db.ErrPlayerAlreadyInGame) {
			ErrorResponse(w, r, http.StatusConflict, "This player is already in the game")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to add player to game")
		}
		responseSent = true
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
			responseSent = true
			return
		}
		if dbUser.DisplayName.Valid && dbUser.DisplayName.String != "" {
			playerDisplayName = dbUser.DisplayName.String
		} else {
			playerDisplayName = dbUser.Username
		}
	}

	if !responseSent {
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
		responseSent = true
	}
}

// Handles removing a player from a game
// Path: /games/{game_id}/players/{game_player_id}
// Method: DELETE
func (gh *GameHandler) HandleRemovePlayerFromGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
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

	tx, txOk := StartTx(ctx, w, r, logger, "Could not remove player")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.DeleteGamePlayer(ctx, tx, gamePlayerID)
	if opErr != nil {
		if errors.Is(opErr, db.ErrGamePlayerNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Player not found in this game")
		} else {
			logger.Error().Err(opErr).Msg("Failed to delete player from game")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not remove player")
		}
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusNoContent, nil, "Player removed successfully")
		responseSent = true
	}
}

// Handles starting a game by updating its status from 'pending' to 'active'
// Path: /games/{game_id}/start
// Method: PUT
func (gh *GameHandler) HandleStartGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
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

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to start game")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	game, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		opErr = errors.New("authorization check failed")
		responseSent = true
		return
	}

	if game.Status != "pending" {
		opErr = fmt.Errorf("game not in pending state: %s", game.Status)
		ErrorResponse(
			w, r, http.StatusConflict,
			fmt.Sprintf("Cannot start game because its status is already '%s'", game.Status),
		)
		responseSent = true
		return
	}

	if !game.StartingDealerGamePlayerID.Valid {
		opErr = errors.New("cannot start game without a starting dealer")
		ErrorResponse(w, r, http.StatusBadRequest, "Game setup is incomplete. Please set a starting dealer")
		responseSent = true
		return
	}

	_, err := db.CreateRound(
		ctx,
		tx,
		gameID,
		game.StartingDealerGamePlayerID.String,
		1, false,
	)
	if err != nil {
		opErr = fmt.Errorf("failed to create the first round: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not create the first round to start the game")
		responseSent = true
		return
	}
	logger.Info().Msg("Successfully created round 1 for the game")

	players, err := db.GetPlayersByGameID(ctx, tx, gameID)
	if err != nil {
		opErr = fmt.Errorf("failed to get players for game start: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve players to start the game")
		responseSent = true
		return
	}
	if len(players) < 2 {
		opErr = errors.New("not enough players to start game")
		ErrorResponse(w, r, http.StatusBadRequest, "Cannot start a game with fewer than 2 players")
		responseSent = true
		return
	}

	if err := db.UpdateGameStatus(ctx, nil, gameID, "active"); err != nil {
		opErr = fmt.Errorf("failed to update game status to active: %w", err)
		logger.Error().Err(opErr).Msg("Failed to update game status to active")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to start the game")
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Game started successfully")
	}
}

// Handles fetching all players for a specific game
// Path: /games/{game_id}/players
// Method: GET
func (gh *GameHandler) HandleGetGamePlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
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
		gameComponent,
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
		gameComponent,
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
		"scorekeeper_user_id":            req.ScorekeeperUserID,
		"starting_dealer_game_player_id": req.StartingDealerGamePlayerID,
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	// --- Validate ---
	dbPlayers, err := db.GetPlayersByGameID(ctx, tx, gameID)
	if err != nil {
		opErr = err
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verify game players")
		responseSent = true
		return
	}

	playerIDSet := make(map[string]bool)
	for _, p := range dbPlayers {
		playerIDSet[p.GamePlayerID] = true
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
		ErrorResponse(
			w, r, http.StatusBadRequest,
			"Selected scorekeeper is not a valid player in this game",
		)
		responseSent = true
		return
	}

	// Validate starting dealer
	if !playerIDSet[req.StartingDealerGamePlayerID] {
		opErr = fmt.Errorf("invalid starting dealer: %s", req.StartingDealerGamePlayerID)
		ErrorResponse(w, r, http.StatusBadRequest, "Selected starting dealer is not in this game")
		responseSent = true
		return
	}

	// Validate seating order
	if len(req.OrderedPlayerIDs) != len(dbPlayers) {
		opErr = fmt.Errorf(
			"player list length mismatch: expected %d, got %d",
			len(dbPlayers), len(req.OrderedPlayerIDs),
		)
		ErrorResponse(w, r, http.StatusBadRequest,
			"Player list mismatch: ensure all players are included in the seating order",
		)
		responseSent = true
		return
	}

	for _, reqPlayerID := range req.OrderedPlayerIDs {
		if !playerIDSet[reqPlayerID] {
			opErr = fmt.Errorf("invalid player ID in seating order: %s", reqPlayerID)
			ErrorResponse(
				w, r, http.StatusBadRequest,
				fmt.Sprintf("Invalid player ID %s found in seating order", reqPlayerID),
			)
			responseSent = true
			return
		}
	}

	finalOrderedPlayerIDs := req.OrderedPlayerIDs
	dealerIndex := -1
	for i, id := range finalOrderedPlayerIDs {
		if id == req.StartingDealerGamePlayerID {
			dealerIndex = i
			break
		}
	}

	if dealerIndex != -1 {
		reordered := append(finalOrderedPlayerIDs[dealerIndex:], finalOrderedPlayerIDs[:dealerIndex]...)
		finalOrderedPlayerIDs = reordered
		logger.Debug().
			Interface("original_order", req.OrderedPlayerIDs).
			Interface("final_order", finalOrderedPlayerIDs).
			Msg("Reordered players based on new starting dealer")
	}

	opErr = db.UpdateGameSettings(
		ctx, tx, gameID,
		req.ScorekeeperUserID,
		req.StartingDealerGamePlayerID,
		finalOrderedPlayerIDs,
	)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save game settings")
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Game settings updated successfully")
		responseSent = true
	}
}

// Handles fetching active games for the authenticated user
// Path: /games/active
// Method: GET
func (gh *GameHandler) HandleGetActiveGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"HandleGetActiveGames",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	dbGames, err := db.GetActiveGamesByUserID(ctx, nil, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get active games from database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve active games")
		return
	}

	apiGames := make([]apiModels.ActiveGameResponse, 0, len(dbGames))
	for _, dbGame := range dbGames {
		var players []apiModels.ActiveGamePlayer
		if err := json.Unmarshal(dbGame.PlayersJSON, &players); err != nil {
			logger.Error().Err(err).Str(l.GameIDKey, dbGame.GameID).Msg("Failed to unmarshal players JSON")
			continue
		}

		apiGame := apiModels.ActiveGameResponse{
			GameID:          dbGame.GameID,
			ScorekeeperName: dbGame.ScorekeeperName.String,
			IsScorekeeper:   dbGame.IsScorekeeper,
			CreatedAt:       dbGame.CreatedAt,
			CurrentRound:    int(dbGame.CurrentRound.Int32),
			Players:         players,
		}
		if dbGame.SessionName.Valid {
			apiGame.SessionName = &dbGame.SessionName.String
		}
		apiGames = append(apiGames, apiGame)
	}

	Respond(w, r, http.StatusOK, apiGames, "Active games retrieved successfully")
}

// Handles fetching a paginated list of a user's game history
// Path: /games/history
// Method: GET
func (gh *GameHandler) HandleGetGameHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		gameComponent,
		"HandleGetGameHistory",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	page, pageSize := GetPaginationParams(r)
	sortBy := QueryParam(r, "sort_by")
	sortOrder := QueryParam(r, "sort_order")
	sessionId := QueryParam(r, "session_id")

	totalCount, err := db.CountUserGameHistory(ctx, nil, userID, sessionId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to count user game history")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve game history")
		return
	}

	dbHistory, err := db.GetUserGameHistory(ctx, nil, userID, sortBy, sortOrder, sessionId, page, pageSize)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user game history from database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve game history")
		return
	}

	apiHistory := make([]apiModels.GameHistoryItem, 0, len(dbHistory))
	for _, h := range dbHistory {
		item := apiModels.GameHistoryItem{
			GameID:            h.GameID,
			GameDate:          h.GameDate,
			FinishingPosition: int(h.FinishingPosition.Int32),
			TotalPoints:       int(h.TotalPoints.Int32),
			RoundsHit:         int(h.RoundsHit.Int32),
			ZeroDifferential:  int(h.ZeroDifferential.Int32),
			TotalPlayers:      int(h.TotalPlayers.Int32),
			ScorekeeperName:   h.ScorekeeperName.String,
		}
		if h.SessionName.Valid {
			item.SessionName = &h.SessionName.String
		}
		apiHistory = append(apiHistory, item)
	}

	pagination := CalculatePagination(totalCount, page, pageSize)

	response := apiModels.PaginatedGameHistoryResponse{
		Games:      apiHistory,
		Pagination: pagination,
	}

	Respond(w, r, http.StatusOK, response, "Game history retrieved successfully")
}
