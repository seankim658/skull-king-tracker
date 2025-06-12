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

const friendshipComponent = "handlers-friendship"

type FriendshipHandler struct {
	Cfg *cf.Config
}

func NewFriendshipHandler(cfg *cf.Config) *FriendshipHandler {
	return &FriendshipHandler{Cfg: cfg}
}

// Handles a user's request to friend another user
func (fh *FriendshipHandler) HandleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleSendFriendRequest",
	)

	requesterID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.RequesterIDKey, requesterID).Logger()

	var req apiModels.SendFriendRequest
	if !ParseJSON(w, r, &req) {
		return
	}
	if !RequireFields(w, r, map[string]string{
		"addressee_id": req.AddresseeID,
	}) {
		return
	}
	if requesterID == req.AddresseeID {
		ErrorResponse(w, r, http.StatusBadRequest, "You cannot send a friend request to yourself")
		return
	}
	logger = logger.With().Str(l.AddresseeIDKey, req.AddresseeID).Logger()

	tx, txOk := StartTx(ctx, w, r, logger, "Failed to send friend request")
	if !txOk {
		return
	}
	var opErr error
	defer func() {
		if p := recover(); p != nil {
			opErr = fmt.Errorf("panic recovered: %v", p)
			logger.Error().Err(opErr).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic in HandleSendFriendRequest")
		}
		if opErr != nil {
			logger.Warn().Err(opErr).Msg("Rolling back transaction")
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Error().Err(rbErr).Msg("Transaction rollback failed")
			}
		}
	}()

	friendshipID, opErr := db.CreateFriendship(ctx, tx, requesterID, req.AddresseeID)
	if opErr != nil {
		if errors.Is(opErr, db.ErrFriendshipAlreadyExists) {
			ErrorResponse(w, r, http.StatusConflict, "A pending or accepted friendship already exists with this user")
		} else if errors.Is(opErr, db.ErrFriendshipBlocked) {
			ErrorResponse(w, r, http.StatusForbidden, "Cannot send a friend request to this user")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to send friend request")
		}
		return
	}

	actor, err := db.GetUserByID(ctx, tx, requesterID)
	if err != nil {
		opErr = fmt.Errorf("failed to get actor for notification: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not send friend request")
		return
	}

	actorDisplayName := actor.Username
	if actor.DisplayName.Valid {
		actorDisplayName = actor.DisplayName.String
	}
	message := fmt.Sprintf("%s wants to be your friend", actorDisplayName)
	_, opErr = db.CreateNotification(
		ctx,
		tx,
		req.AddresseeID,
		requesterID,
		apiModels.NotificationTypeFriendRequest,
		message,
		&friendshipID,
	)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Friend reqeust sent, but failed to create a notification")
		return
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for friend request: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save friend request")
		return
	}

	Respond(w, r, http.StatusCreated, nil, "Friendship request sent successfully")
}

// Handles accepting or declining a friend request
func (fh *FriendshipHandler) HandleRespondToFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleRespondToFriendRequest",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	friendshipID, ok := PathVar(w, r, "friendship_id")
	if !ok {
		return
	}

	var req apiModels.RespondToFriendRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	var newStatus string
	if req.Response == "accept" {
		newStatus = "accepted"
	} else {
		// We're just assuming anything not explicitly an "accept" will decline the request
		newStatus = "declined"
	}

	friendship, err := db.GetFriendshipByID(ctx, nil, friendshipID)
	if err != nil {
		if errors.Is(err, db.ErrFriendhipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Friend request not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process request")
		}
		return
	}

	if friendship.AddresseeID != userID {
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to response to this friend request")
		return
	}

	if friendship.Status != "pending" {
		ErrorResponse(w, r, http.StatusConflict, "This friend request has already been actioned")
		return
	}

	err = db.UpdateFriendshipStatus(ctx, nil, friendshipID, newStatus)
	if err != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to respond to friend request")
		return
	}

	Respond(w, r, http.StatusOK, nil, fmt.Sprintf("Friend request %s", newStatus))
}
