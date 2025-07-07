package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/seankim658/skullking/internal/sse"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const scoringComponent = "handlers-scoring"

type ScoringHandler struct {
	Cfg    *cf.Config
	SSEHub *sse.Hub
}

func NewScoringHandler(cfg *cf.Config, sseHub *sse.Hub) *ScoringHandler {
	return &ScoringHandler{Cfg: cfg, SSEHub: sseHub}
}

// Handles retrieving the scorecard state
func (sh *ScoringHandler) HandleGetScorecardState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"HandleGetScorecardState",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Logger()

	userID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}

	isPlayer, err := db.IsUserInGame(ctx, nil, userID, gameID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verify game access")
		return
	}

	game, err := db.GetGameByID(ctx, nil, gameID)
	if err != nil {
		if errors.Is(err, db.ErrGameNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Game not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve game")
		}
		return
	}

	if !isPlayer && game.CreatedByUserID != userID {
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to view this scorecard")
		return
	}

	scorecardData, err := db.GetScorecardState(ctx, nil, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get scorecard state from DB")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve scorecard")
		return
	}

	apiResponse := modelConverters.DBScorecardToAPIResponse(scorecardData)
	Respond(w, r, http.StatusOK, apiResponse, "Scorecard state retrieved successfully")
}

// Handles submitting round bids
func (sh *ScoringHandler) HandleSubmitBids(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"HandleSubmitBids",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	roundNumber, ok := PathVarInt(w, r, "round_number")
	if !ok {
		return
	}

	userID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}

	_, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		return
	}

	var req apiModels.SubmitBidsRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	gamePlayers, err := db.GetPlayersByGameID(ctx, nil, gameID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not verify game players for validation")
		return
	}
	if len(req.Bids) != len(gamePlayers) {
		msg := fmt.Sprintf(
			"Bid payload is incomplete. Expected %d player bids, but received %d",
			len(gamePlayers), len(req.Bids),
		)
		ErrorResponse(w, r, http.StatusBadRequest, msg)
		return
	}

	playerIDSet := make(map[string]bool)
	for _, p := range gamePlayers {
		playerIDSet[p.GamePlayerID] = true
	}
	for _, b := range req.Bids {
		if !playerIDSet[b.GamePlayerID] {
			msg := fmt.Sprintf("Bid submitted for a player (ID: %s) who is not in this game", b.GamePlayerID)
			ErrorResponse(w, r, http.StatusBadRequest, msg)
			return
		}
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to submit bids")
	if !txOk {
		return
	}

	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	currentRound, err := db.GetCurrentRoundInfo(ctx, tx, gameID)
	if err != nil {
		opErr = err
		if errors.Is(err, db.ErrNoRoundsFound) {
			ErrorResponse(w, r, http.StatusNotFound, "No active round found for this game")
			responseSent = true
			return
		}
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve current round information")
		responseSent = true
		return
	}

	if currentRound.RoundNumber != roundNumber {
		opErr = fmt.Errorf("round number mismatch: expected %d, got %d", currentRound.RoundNumber, roundNumber)
		ErrorResponse(w, r, http.StatusConflict, opErr.Error())
		responseSent = true
		return
	}

	if currentRound.Status != "bidding" {
		opErr = fmt.Errorf("cannot submit bids for a round with status '%s'", currentRound.Status)
		ErrorResponse(w, r, http.StatusConflict, opErr.Error())
		responseSent = true
		return
	}

	dbBids := make([]dbModels.PlayerBidData, len(req.Bids))
	for i, apiBid := range req.Bids {
		dbBids[i] = dbModels.PlayerBidData{
			GamePlayerID: apiBid.GamePlayerID,
			BidAmount:    apiBid.BidAmount,
		}
	}

	opErr = db.SubmitBidsAndUpdateRoundStatus(ctx, tx, currentRound.RoundID, dbBids)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save bids to the database")
		responseSent = true
		return
	}

	go broadcastScorecardUpdate(sh.SSEHub, gameID, logger)

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Bids submitted successfully")
		responseSent = true
	}
}

// Handles submitting the tricks taken and bonus points for a round
func (sh *ScoringHandler) HandleSubmitTricks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		scoringComponent,
		"HandleSubmitTricks",
	)

	gameID, ok := PathVar(w, r, "game_id")
	if !ok {
		return
	}
	roundNumber, ok := PathVarInt(w, r, "round_number")
	if !ok {
		return
	}
	logger = logger.With().Str(l.GameIDKey, gameID).Int(l.RoundKey, roundNumber).Logger()

	userID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}

	_, authorized := CheckGameAccessAndScorekeeper(ctx, w, r, gameID, userID, logger)
	if !authorized {
		return
	}

	var req apiModels.SubmitTricksRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to submit scores")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	currentRound, err := db.GetCurrentRoundInfo(ctx, tx, gameID)
	if err != nil {
		opErr = err
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve current round information")
		responseSent = true
		return
	}

	if currentRound.Status != "playing" {
		opErr = fmt.Errorf("cannot submit scores for a round with status '%s'", currentRound.Status)
		ErrorResponse(w, r, http.StatusConflict, opErr.Error())
		responseSent = true
		return
	}

	scores := make([]dbModels.PlayerScoreData, len(req.Tricks))
	for i, tricksData := range req.Tricks {
		scores[i] = dbModels.PlayerScoreData{
			GamePlayerID: tricksData.GamePlayerID,
			TricksTaken:  tricksData.TricksTaken,
			BonusPoints:  tricksData.BonusPoints,
		}
	}

	opErr = db.SubmitScoresAndUpdateRound(ctx, tx, currentRound.RoundID, scores)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save scores to the database")
		responseSent = true
		return
	}

  if currentRound.RoundNumber < 10 {
    logger.Info().Msg("Round completed, creating next round")

    gamePlayers, err := db.GetPlayersByGameID(ctx, tx, gameID)
    if err != nil {
      opErr = fmt.Errorf("failed to get game players for dealer rotation: %w", err)
      ErrorResponse(w, r, http.StatusInternalServerError, "Could not prepare the next round")
      responseSent = true
      return
    }

    if len(gamePlayers) == 0 {
      opErr = errors.New("no players found in game to determine next dealer")
      ErrorResponse(w, r, http.StatusInternalServerError, "Could not prepare the next round: no players found")
      responseSent = true
      return
    }

    currentDealerIndex := -1
    for i, p := range gamePlayers {
      if p.GamePlayerID == currentRound.DealerGamePlayerID {
        currentDealerIndex = i
        break
      }
    }

    if currentDealerIndex == -1 {
      opErr = errors.New("could not find current dealer in player list")
      ErrorResponse(w, r, http.StatusInternalServerError, "Could not prepare the next round: dealer not found")
      responseSent = true
      return
    }

    nextDealerIndex := (currentDealerIndex + 1) % len(gamePlayers)
    nextDealerID := gamePlayers[nextDealerIndex].GamePlayerID
    nextRoundNumber := currentRound.RoundNumber + 1

    _, err = db.CreateRound(ctx, tx, gameID, nextDealerID, nextRoundNumber, false)
    if err != nil {
      opErr = fmt.Errorf("failed to create round %d: %w", nextRoundNumber, err)
      ErrorResponse(w, r, http.StatusInternalServerError, "Failed to start the next round")
      responseSent = true
      return
    }
    logger.Info().Int(l.RoundKey, nextRoundNumber).Msg("Next round created successfully")
  } else {
    logger.Info().Msg("Final round completed, marking game as finished")
    if err := db.UpdateGameStatus(ctx, tx, gameID, "completed"); err != nil {
      opErr = fmt.Errorf("failed to mark game as completed: %w", err)
      ErrorResponse(w, r, http.StatusInternalServerError, "Failed to finalize the game")
      responseSent = true
      return
    }
  }

	go broadcastScorecardUpdate(sh.SSEHub, gameID, logger)

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Scores submitted successfully")
		responseSent = true
	}
}

// Helper function to broadcast scorecard updates via SSE
func broadcastScorecardUpdate(sseHub *sse.Hub, gameID string, logger zerolog.Logger) {
	broadcastCtx := l.NewContextWithLogger(context.Background(), logger)

	scorecardData, err := db.GetScorecardState(broadcastCtx, nil, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("SSE broadcast failed to get latest scorecard state")
		return
	}

	apiResponse := modelConverters.DBScorecardToAPIResponse(scorecardData)

	ssePayload := apiModels.SSEEvent{
		Event:   "scorecard_updated",
		Payload: apiResponse,
	}

	jsonPayload, err := json.Marshal(ssePayload)
	if err != nil {
		logger.Error().Err(err).Msg("SSE broadcast failed to marshal payload")
		return
	}

	players, err := db.GetPlayersByGameID(broadcastCtx, nil, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("SSE broadcast failed to get players for game")
	}

	broadcastMessage := string(jsonPayload)
	broadcastCount := 0
	for _, player := range players {
		if player.UserID.Valid {
			sseHub.Broadcast(player.UserID.String, broadcastMessage)
			broadcastCount++
		}
	}
	logger.Info().Str(l.GameIDKey, gameID).Msgf("Broadcast 'scorecard_updated' to %d players", broadcastCount)
}
