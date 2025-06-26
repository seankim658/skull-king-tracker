package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
)

const sessionHandlerComponent = "handlers-session"

type SessionHandler struct {
	Cfg *cf.Config
}

func NewSessionHandler(cfg *cf.Config) *SessionHandler {
	return &SessionHandler{Cfg: cfg}
}

// Retrieves active game sessions for the authenticated user
// Path: /sessions/active
// Method: GET
func (sh *SessionHandler) HandleGetActiveSessionsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionHandlerComponent,
		"HandleGetActiveSessionsForUser",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	dbSessionsWithActivity, err := db.GetActiveSessionsByUserID(ctx, nil, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve active sessions for user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve active sessions")
		return
	}

	apiSessions := make([]apiModels.ActiveSessionResponse, 0, len(dbSessionsWithActivity))
	for _, dbSess := range dbSessionsWithActivity {
		apiSess := apiModels.ActiveSessionResponse{
			SessionID:      dbSess.SessionID,
			Status:         dbSess.Status,
			HasActiveGame:  dbSess.HasActiveGame,
			HasPendingGame: dbSess.HasPendingGame,
			CreatedAt:      dbSess.CreatedAt,
			UpdatedAt:      dbSess.UpdatedAt,
		}
		if dbSess.SessionName.Valid {
			apiSess.SessionName = &dbSess.SessionName.String
		}
		if dbSess.CompletedAt.Valid {
			apiSess.CompletedAt = &dbSess.CompletedAt.Time
		}
		if dbSess.CreatorName.Valid {
			apiSess.CreatorName = &dbSess.CreatorName.String
		}
		apiSessions = append(apiSessions, apiSess)
	}

	Respond(w, r, http.StatusOK, apiSessions, "Successfully retrieved active sessions")
}

// Marks a session as completed
// Path: /api/sessions/{session_id}/complete
// Method: PUT
func (sh *SessionHandler) HandleCompleteSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionHandlerComponent,
		"HandleCompleteSession",
	)

	sessionID, ok := PathVar(w, r, "session_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.SessionIDKey, sessionID).Logger()

	userID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	session, err := db.GetGameSessionByID(ctx, nil, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrSessionNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Session not found")
		} else {
			logger.Error().Err(err).Msg("Failed to get session for auth check")
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to verify session access")
		}
		return
	}
	if !session.CreatedByUserID.Valid || session.CreatedByUserID.String != userID {
		logger.Warn().
			Str("creator_id", session.CreatedByUserID.String).
			Msg("User is not the creator of the session")
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to complete this session")
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to complete session")
	if !txOk {
		return
	}

	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	completedTime := sql.NullTime{Time: time.Now(), Valid: true}
	opErr = db.UpdateSessionStatus(ctx, tx, sessionID, "completed", completedTime)
	if opErr != nil {
		if errors.Is(opErr, db.ErrSessionNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Session not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update session status")
		}
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Session marked as completed sucessfully")
		responseSent = true
	}
}

// Handles fetching the details of a single session, including its games and user-specific statistics
// Path: /sessions/{session_id}
// Method: GET
func (sh *SessionHandler) HandleGetSessionDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionHandlerComponent,
		"HandleGetSessionDetails",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	sessionID, ok := PathVar(w, r, "session_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.SessionIDKey, sessionID).Logger()

	participated, err := db.CheckUserParticipatedInSession(ctx, nil, userID, sessionID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check user participation for auth")
		return
	}
	if !participated {
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to view this session's details")
		return
	}

	dbSession, err := db.GetGameSessionByID(ctx, nil, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrSessionNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Session not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve session details")
		}
		return
	}

	dbGames, err := db.GetGamesBySessionID(ctx, nil, sessionID, userID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve games for the session")
		return
	}

	dbUserStats, err := db.GetUserSessionStats(ctx, nil, userID, sessionID)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve user stats for the session")
		return
	}

	apiGames := make([]apiModels.SessionGame, len(dbGames))
	for i, dbGame := range dbGames {
		apiGames[i] = apiModels.SessionGame{
			GameID:        dbGame.GameID,
			Status:        dbGame.Status,
			CreatedAt:     dbGame.CreatedAt,
			IsScorekeeper: dbGame.IsViewerScorekeeper,
		}
		if dbGame.CompletedAt.Valid {
			completedAtStr := dbGame.CompletedAt.Time.String()
			apiGames[i].CompletedAt = &completedAtStr
		}
		if dbGame.WinningPlayer.Valid {
			apiGames[i].WinningPlayer = &dbGame.WinningPlayer.String
		}
		if dbGame.ScorekeeperName.Valid {
			apiGames[i].ScorekeeperName = &dbGame.ScorekeeperName.String
		}
	}

	apiResponse := apiModels.SessionDetailResponse{
		SessionID: dbSession.SessionID,
		Status:    dbSession.Status,
		Games:     apiGames,
		UserSummary: apiModels.SessionUserSummary{
			TotalGames: dbUserStats.TotalGamesPlayed,
			Wins:       dbUserStats.TotalWins,
		},
	}
	if dbSession.SessionName.Valid {
		apiResponse.SessionName = &dbSession.SessionName.String
	}

	Respond(w, r, http.StatusOK, apiResponse, "Session details retrived successfully")
}
