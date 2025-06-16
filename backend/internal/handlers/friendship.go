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

	addresseeID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, addresseeID).Logger()

	friendshipID, ok := PathVar(w, r, "friendship_id")
	if !ok {
		return
	}

	var req apiModels.RespondToFriendRequest
	if !ParseJSON(w, r, &req) {
		return
	}

	tx, txOk := StartTx(ctx, w, r, logger, "Clould not process friend request transaction")
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

	friendship, err := db.GetFriendshipByID(ctx, tx, friendshipID)
	if err != nil {
		opErr = err
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Friend request not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process request")
		}
		return
	}

	if friendship.AddresseeID != addresseeID {
		opErr = errors.New("user not authorized to respond to this friend request")
		ErrorResponse(w, r, http.StatusForbidden, "You are not authorized to response to this friend request")
		return
	}

	if friendship.Status != "pending" {
		opErr = errors.New("friend request already actioned")
		ErrorResponse(w, r, http.StatusConflict, "This friend request has already been actioned")
		return
	}

	var newStatus string
	if req.Response == "accept" {
		newStatus = "accepted"
	} else {
		// We're just assuming anything not explicitly an "accept" will decline the request
		newStatus = "declined"
	}
	logger = logger.With().Str("response_action", newStatus).Logger()

	opErr = db.UpdateFriendshipStatus(ctx, tx, friendshipID, newStatus)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to respond to friend request")
		return
	}

	opErr = db.DeleteNotificationByFriendshipID(ctx, tx, friendshipID, apiModels.NotificationTypeFriendRequest)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to delete original friend request notification, but proceeding")
		opErr = nil
	}

	if newStatus == "accepted" {
		actor, userErr := db.GetUserByID(ctx, tx, addresseeID)
		if userErr != nil {
			opErr = fmt.Errorf("failed to get actor for accepted notificaiton: %w", userErr)
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not send acceptance notification")
			return
		}

		actorDisplayName := actor.Username
		if actor.DisplayName.Valid {
			actorDisplayName = actor.DisplayName.String
		}
		message := fmt.Sprintf("%s accepted your friend request", actorDisplayName)

		_, opErr = db.CreateNotification(
			ctx,
			tx,
			friendship.RequesterID,
			addresseeID,
			"friend_accepted",
			message,
			&friendshipID,
		)
		if opErr != nil {
			logger.Error().Err(opErr).Msg("Failed to create 'friend_accepted' notification")
			opErr = nil
		}
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for friend response: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save friend request")
		return
	}

	Respond(w, r, http.StatusOK, nil, fmt.Sprintf("Friend request %s", newStatus))
}

// Handles removing a friendship
func (fh *FriendshipHandler) HandleUnfriend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleUnfriend",
	)

	removerID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	friendToRemoveID, ok := PathVar(w, r, "user_id")
	if !ok {
		return
	}

	logger = logger.With().
		Str(l.UserIDKey, removerID).
		Str("friend_to_remove", friendToRemoveID).
		Logger()

		// TODO : we should probably create a transaction for this
	err := db.DeleteFriendship(ctx, nil, removerID, friendToRemoveID)
	if err != nil {
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Friendship not found")
		} else {
			logger.Error().Err(err).Msg("Failed to unfriend user")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process unfriend request")
		}
		return
	}

	Respond(w, r, http.StatusOK, nil, "Friendship removed successfully")
}

// Allows a user to cancel a request they have sent
func (fh *FriendshipHandler) HandleCancelFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleCancelFriendRequest",
	)

	requesterID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	addresseeID, ok := PathVar(w, r, "addressee_id")
	if !ok {
		return
	}
	logger = logger.With().Str(l.RequesterIDKey, requesterID).Str(l.AddresseeIDKey, addresseeID).Logger()

	tx, txOk := StartTx(ctx, w, r, logger, "Could not process request cancellation")
	if !txOk {
		return
	}
	var opErr error
	defer func() {
		if p := recover(); p != nil {
			opErr = fmt.Errorf("panic recovered: %v", p)
			logger.Error().Err(opErr).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic in HandleCancelFriendRequest")
		}
		if opErr != nil {
			logger.Error().Err(opErr).Msg("Rolling back transaction")
			_ = tx.Rollback()
		}
	}()

	pendingFriendship, err := db.GetPendingFriendshipByUsers(ctx, tx, requesterID, addresseeID)
	if err != nil {
		opErr = err
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "No pending friend request found to cancel")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Error finding friend request")
		}
		return
	}

	opErr = db.DeleteNotificationByFriendshipID(ctx, tx, pendingFriendship.FriendshipID, apiModels.NotificationTypeFriendRequest)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to delete notification for canceled request")
		opErr = nil
	}

	opErr = db.DeleteFriendshipByID(ctx, tx, pendingFriendship.FriendshipID)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to cancel friend request")
		return
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transaction for friend request cancellation: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save cancellation")
		return
	}
	Respond(w, r, http.StatusOK, nil, "Friend request cancelled")
}

// Handles blocking a user
func (fh *FriendshipHandler) HandleBlockUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleBlockUser",
	)

	blockerID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	userToBlockID, ok := PathVar(w, r, "user_id_to_block")
	if !ok {
		return
	}
	if blockerID == userToBlockID {
		ErrorResponse(w, r, http.StatusBadRequest, "Cannot block yourself")
		return
	}
	logger = logger.With().Str(l.BlockerIDKey, blockerID).Str(l.BlockedIDKey, userToBlockID).Logger()

	tx, txOk := StartTx(ctx, w, r, logger, "Could not process block request")
	if !txOk {
		return
	}
	var opErr error
	defer func() {
		if p := recover(); p != nil {
			opErr = fmt.Errorf("panic recovered: %v", p)
			logger.Error().Err(opErr).Bytes(l.StackTraceKey, debug.Stack()).Msg("Panic in HandleBlockUser")
		}
		if opErr != nil {
			logger.Error().Err(opErr).Msg("Rolling back transaction")
			_ = tx.Rollback()
		}
	}()

	opErr = db.BlockUser(ctx, tx, blockerID, userToBlockID)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to block user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not process block request")
		return
	}

	opErr = db.DeleteFriendRequestNotification(ctx, tx, blockerID, userToBlockID)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to clean up pending notification during block, proceeding")
		opErr = nil
	}

	if err := tx.Commit(); err != nil {
		opErr = fmt.Errorf("failed to commit transcation for blocking user: %w", err)
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to save block action")
		return
	}

	Respond(w, r, http.StatusOK, nil, "User blocked successfully")
}

// Handles unblocking a user
func (fh *FriendshipHandler) HandleUnblockUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleUnblockUser",
	)

	unblockerID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	userToUnblockID, ok := PathVar(w, r, "user_id_to_unblock")
	if !ok {
		return
	}
	logger = logger.With().
		Str(l.UnblockerIDKey, unblockerID).
		Str(l.UnblockedIDKey, userToUnblockID).
		Logger()

	err := db.UnblockUser(ctx, nil, unblockerID, userToUnblockID)
	if err != nil {
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "No block record found for this user")
		} else {
			logger.Error().Err(err).Msg("Failed to unblock user")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process unblock request")
		}
		return
	}

	Respond(w, r, http.StatusOK, nil, "User unblocked successfully")
}
