package handlers

import (
	"fmt"
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	convert "github.com/seankim658/skullking/internal/models/convert"
	"github.com/seankim658/skullking/internal/sse"
)

const notificationComponent = "handlers-notification"

type NotificationHandler struct {
	Cfg    *cf.Config
	SSEHub *sse.Hub
}

func NewNotificationHandler(cfg *cf.Config, sseHub *sse.Hub) *NotificationHandler {
	return &NotificationHandler{Cfg: cfg, SSEHub: sseHub}
}

// Establishes an SSE connection for real-time notifications
// Path: /notifications/events
// Method: GET
func (nh *NotificationHandler) HandleNotificationStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"HandleNotificationStream",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}
	logger = logger.With().Str(l.UserIDKey, userID).Logger()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error().Msg("Streaming unsupported - no flusher available")
		ErrorResponse(w, r, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "event: connected\ndata: Connection established\n\n")
	flusher.Flush()

	messageChan := make(chan string)
	nh.SSEHub.AddClient(userID, messageChan)
	logger.Info().Msg("SSE client connected and registered")

	defer func() {
		nh.SSEHub.RemoveClient(userID)
		close(messageChan)
		logger.Info().Msg("SSE client disconnected and unregistered")
	}()

	done := make(chan bool)
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			logger.Info().Msg("SSE client context cancelled")
			return
		case <-r.Context().Done():
			logger.Info().Msg("SSE client request context cancelled")
			return
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case message, open := <-messageChan:
			if !open {
				logger.Info().Msg("SSE message channel closed")
				return
			}

			_, err := fmt.Fprintf(w, "data: %s\n\n", message)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to write SSE message")
				return
			}

			flusher.Flush()
			logger.Debug().Str("message", message).Msg("SSE message sent")
		}
	}
}

// Retrieves the user notifications
// Path: /notifications
// Method: GET
func (nh *NotificationHandler) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"HandleGetNotifications",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	limit, limitOk := QueryParamInt(r, "limit")
	if !limitOk {
		limit = 25
	}
	logger = logger.With().Int(l.LimitKey, limit).Str(l.UserIDKey, userID).Logger()

	dbNotifications, err := db.GetNotificationByUserID(ctx, nil, userID, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve notifications for user")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not retrieve notifications")
		return
	}

	apiNotifications := convert.DBNotificationsToAPIResponse(dbNotifications)
	Respond(w, r, http.StatusOK, apiNotifications, "Notifications retrieved successfully")
}

// Marks a single notification as read
// Path: /notifications/{notification_id}/read
// Method: PUT
func (nh *NotificationHandler) HandleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"HandleMarkNotificationRead",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	notificationID, ok := PathVar(w, r, "notification_id")
	if !ok {
		return
	}

	err := db.UpdateNotificationReadStatus(ctx, nil, notificationID, userID, true)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to mark notification")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not mark notification as read")
		return
	}

	Respond(w, r, http.StatusNoContent, nil, "Notification marked as read")
}

// Marks a single notification as unread
// Path: /notifications/{notification_id}/read
// Method: DELETE
func (nh *NotificationHandler) HandleMarkNotificationUnread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"HandleMarkNotificationUnread",
	)

	userID, ok := GetAuthenticatedUserIDFromSession(w, r, logger)
	if !ok {
		return
	}

	notificationID, ok := PathVar(w, r, "notification_id")
	if !ok {
		return
	}

	err := db.UpdateNotificationReadStatus(ctx, nil, notificationID, userID, false)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to mark notification as unread")
		ErrorResponse(w, r, http.StatusInternalServerError, "Could not mark notification as unread")
		return
	}

	Respond(w, r, http.StatusNoContent, nil, "Notification marked as unread")
}
