package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	apiModels "github.com/seankim658/skullking/internal/models/api"
	modelConverters "github.com/seankim658/skullking/internal/models/convert"
	"github.com/seankim658/skullking/internal/sse"
)

const friendshipComponent = "handlers-friendship"

type FriendshipHandler struct {
	Cfg    *cf.Config
	SSEHub *sse.Hub
}

func NewFriendshipHandler(cfg *cf.Config, sseHub *sse.Hub) *FriendshipHandler {
	return &FriendshipHandler{Cfg: cfg, SSEHub: sseHub}
}

// Handles a user's request to friend another user
// Path: /friends/request
// Method: POST
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	createdNotification, err := db.SendFriendRequest(ctx, tx, requesterID, req.AddresseeID)
	opErr = err
	if opErr != nil {
		if errors.Is(opErr, db.ErrFriendshipAlreadyExists) {
			ErrorResponse(w, r, http.StatusConflict, "A pending or accepted friendship already exists with this user")
		} else if errors.Is(opErr, db.ErrFriendshipBlocked) {
			ErrorResponse(w, r, http.StatusForbidden, "Cannot send a friend request to this user")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Failed to send friend request")
		}
		responseSent = true
		return
	}

	go broadcastNotificationEvent(fh.SSEHub, createdNotification, logger)

	if !responseSent {
		Respond(w, r, http.StatusCreated, nil, "Friendship request sent successfully")
		responseSent = true
	}
}

// Handles accepting or declining a friend request
// Path: /friends/request/{friendship_id}
// Method: PUT
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

	tx, txOk := StartTx(ctx, w, r, logger, "Could not process friend request transaction")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	friendship, err := db.GetFriendshipByID(ctx, tx, friendshipID)
	if err != nil {
		opErr = err
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Friend request not found")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process request")
		}
		responseSent = true
		return
	}

	if friendship.AddresseeID != addresseeID || friendship.Status != "pending" {
		opErr = errors.New("user not authorized or request not pending")
		ErrorResponse(
			w, r, http.StatusForbidden,
			"You are not authorized to respond to this friend request",
		)
		responseSent = true
		return
	}

	newStatus := "declined"
	if req.Response == "accept" {
		newStatus = "accepted"
	}
	logger = logger.With().Str("response_action", newStatus).Logger()

	acceptedNotification, err := db.RespondToFriendRequest(ctx, tx, friendshipID, newStatus)
	opErr = err
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to respond to friend request")
		responseSent = true
		return
	}

	if acceptedNotification != nil {
		go broadcastNotificationEvent(fh.SSEHub, acceptedNotification, logger)
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, fmt.Sprintf("Friend request %s", newStatus))
		responseSent = true
	}
}

// Handles removing a friendship
// Path: /friends/{user_id}
// Method: DELETE
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

	tx, txOk := StartTx(ctx, w, r, logger, "Could not processs unfriend request")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.DeleteFriendship(ctx, tx, removerID, friendToRemoveID)
	if opErr != nil {
		if errors.Is(opErr, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "Friendship not found")
		} else {
			logger.Error().Err(opErr).Msg("Failed to unfriend user")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process unfriend request")
		}
		responseSent = true
		return
	}

	Respond(w, r, http.StatusOK, nil, "Friendship removed successfully")
	responseSent = true
}

// Allows a user to cancel a request they have sent
// Path: /friends/request/cancel/{addressee_id}
// Method: DELETE
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	// 1. Find the pending friendship to get its ID
	pendingFriendship, err := db.GetPendingFriendshipByUsers(ctx, tx, requesterID, addresseeID)
	if err != nil {
		opErr = err
		if errors.Is(err, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "No pending friend request found to cancel")
		} else {
			ErrorResponse(w, r, http.StatusInternalServerError, "Error finding friend request")
		}
		responseSent = true
		return
	}

	// 2. Find the notification associated with this request to get its ID for the SSE event
	notificationToCancel, err := db.GetNotificationByUsersAndType(
		ctx, tx, addresseeID, requesterID,
		apiModels.NotificationTypeFriendRequest,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error().Err(err).Msg("Could not find notifiaction to broadcast")
	}

	// 3. Broadcast deletion via SSE
	if notificationToCancel != nil {
		ssePayload := apiModels.SSEEvent{
			Event: "notification_deleted",
			Payload: apiModels.SSEDeletedNotificationPayload{
				NotificationID: notificationToCancel.NotificationID,
			},
		}
		jsonPayload, jsonErr := json.Marshal(ssePayload)
		if jsonErr != nil {
			logger.Error().Err(jsonErr).Msg("Failed to marshal SSE deletion event payload")
		} else {
			fh.SSEHub.Broadcast(addresseeID, string(jsonPayload))
			logger.Info().Str(l.RecipientIDKey, addresseeID).Msg("Broadcasted 'notification_deleted' event via SSE")
		}
	}

	// 4. Delete the notification from the database
	if err := db.DeleteNotificationByFriendshipID(
		ctx, tx, pendingFriendship.FriendshipID,
		apiModels.NotificationTypeFriendRequest,
	); err != nil {
		logger.Error().Err(opErr).Msg("Failed to delete notification for canceled request")
	}

	// 5. Delete the friendship record
	opErr = db.DeleteFriendshipByID(ctx, tx, pendingFriendship.FriendshipID)
	if opErr != nil {
		ErrorResponse(w, r, http.StatusInternalServerError, "Failed to cancel friend request")
		responseSent = true
		return
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "Friend request cancelled")
		responseSent = true
	}
}

// Handles blocking a user
// Path: /friends/block/{user_id_to_block}
// Method: POST
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
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.BlockUser(ctx, tx, blockerID, userToBlockID)
	if opErr != nil {
		logger.Error().Err(opErr).Msg("Failed to block user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not process block request")
		responseSent = true
		return
	}

	if err := db.DeleteFriendRequestNotification(ctx, tx, blockerID, userToBlockID); err != nil {
		logger.Error().Err(opErr).Msg("Failed to clean up pending notification during block, proceeding")
	}

	if !responseSent {
		Respond(w, r, http.StatusOK, nil, "User blocked successfully")
		responseSent = true
	}
}

// Handles unblocking a user
// Path: /friends/block/{user_id_to_block}
// Method: DELETE
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

	tx, txOk := StartTx(ctx, w, r, logger, "Could not process unblock request")
	if !txOk {
		return
	}
	var opErr error
	var responseSent bool
	defer ManageTransaction(tx, &opErr, logger, &responseSent)

	opErr = db.UnblockUser(ctx, tx, unblockerID, userToUnblockID)
	if opErr != nil {
		if errors.Is(opErr, db.ErrFriendshipNotFound) {
			ErrorResponse(w, r, http.StatusNotFound, "No block record found for this user")
		} else {
			logger.Error().Err(opErr).Msg("Failed to unblock user")
			ErrorResponse(w, r, http.StatusInternalServerError, "Could not process unblock request")
		}
		responseSent = true
		return
	}

	Respond(w, r, http.StatusOK, nil, "User unblocked successfully")
	responseSent = true
}

// Handles fetching the authenticated user's list of friends
// Path: /friends
// Method: GET
func (fh *FriendshipHandler) HandleGetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		friendshipComponent,
		"HandleGetFriends",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	dbFriends, err := db.GetFriendsByUserID(ctx, nil, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get friends from database")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve your friends list")
		return
	}

	apiFriends := make(apiModels.UserSearchResponse, 0, len(dbFriends))
	for _, dbFriend := range dbFriends {
		apiFriend, convErr := modelConverters.DBUserSearchResultToAPISearchItem(&dbFriend)
		if convErr != nil {
			logger.Error().Err(convErr).Msg("Failed to convert DB friend to API model")
			continue
		}
		apiFriends = append(apiFriends, *apiFriend)
	}

	Respond(w, r, http.StatusOK, apiFriends, "Friends list retrived successfully")
}
