package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"

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
		logger.Error().Err(err).Msg("Failed to retrive active sessions for user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to retrieve active sessions")
		return
	}

	apiSessions := make([]apiModels.ActiveSessionResponse, 0, len(dbSessionsWithActivity))
	for _, dbSess := range dbSessionsWithActivity {
		apiSess := apiModels.ActiveSessionResponse{
			SessionID:     dbSess.SessionID,
			Status:        dbSess.Status,
			HasActiveGame: dbSess.HasActiveGame,
			CreatedAt:     dbSess.CreatedAt,
			UpdatedAt:     dbSess.UpdatedAt,
		}
		if dbSess.SessionName.Valid {
			apiSess.SessionName = &dbSess.SessionName.String
		}
		if dbSess.CompletedAt.Valid {
			apiSess.CompletedAt = &dbSess.CompletedAt.Time
		}
		apiSessions = append(apiSessions, apiSess)
	}

	Respond(w, r, http.StatusOK, apiSessions, "Successfully retrieved active sessions")
}

// Marks a session as completed
func (sh *SessionHandler) HandleCompleteSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		sessionHandlerComponent,
		"HandleCompleteSession",
	)

	vars := mux.Vars(r)
	sessionID, found := vars["session_id"]
	if !found || sessionID == "" {
		ErrorResponse(w, r, http.StatusBadRequest, "Session ID is required in the URL path")
		return
	}
	logger = logger.With().Str(l.SessionIDKey, sessionID).Logger()

	userID, authOk := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !authOk {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	// TODO : Check if the current user can complete this session

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to complete session")
	if !txOk {
		return
	}

	var opErr error
	defer func() {
		if p := recover(); p != nil {
			logger.Error().Interface(l.PanicKey, p).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic recovered")
			_ = tx.Rollback()
		} else if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction due to error in handler logic")
			_ = tx.Rollback()
		}
	}()

	completedTime := sql.NullTime{Time: time.Now(), Valid: true}
	opErr = db.UpdateSessionStatus(ctx, tx, sessionID, "completed", completedTime)
	if opErr != nil {
		if errors.Is(opErr, db.ErrSessionNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Session not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to update session status")
		}
		return
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for completing session: %w", err)
		logger.Error().Err(opErr).Msg("Transaction commit failed")
		return
	}

	Respond(w, r, http.StatusOK, nil, "Session marked as completed sucessfully")
}
