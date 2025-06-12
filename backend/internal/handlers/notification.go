package handlers

import (
	"net/http"

	cf "github.com/seankim658/skullking/internal/config"
	db "github.com/seankim658/skullking/internal/database"
	l "github.com/seankim658/skullking/internal/logger"
	convert "github.com/seankim658/skullking/internal/models/convert"
)

const notificationComponent = "handlers-notification"

type NotificationHandler struct {
	Cfg *cf.Config
}

func NewNotificationHandler(cfg *cf.Config) *NotificationHandler {
	return &NotificationHandler{Cfg: cfg}
}

// Retrieves the user notifications
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
