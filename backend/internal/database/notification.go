package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	l "github.com/seankim658/skullking/internal/logger"
	dbModels "github.com/seankim658/skullking/internal/models/database"
)

const notificationComponent = "database-notification"

// Inserts a new notification into the database
func CreateNotification(
	ctx context.Context,
	tx *sql.Tx,
	recipientID, actorID, notificationType, message string,
	friendshipID *string,
) (string, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"CreateNotification",
	).With().
		Str(l.RecipientIDKey, recipientID).
		Str(l.ActorIDKey, actorID).
		Str(l.NotificationTypeKey, notificationType).
		Logger()

	newNotificationID := uuid.NewString()
	query := `
  INSERT INTO user_notifications (
    notification_id, recipient_user_id, 
    actor_user_id, type, message, friendship_id
  )
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING notification_id;
  `

	var returnedID string
	err := querier.QueryRowContext(
		ctx,
		query,
		newNotificationID,
		recipientID,
		actorID,
		notificationType,
		message,
		friendshipID,
	).Scan(&returnedID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create notification")
		return "", fmt.Errorf("error creating notification: %w", err)
	}
	logger.Info().
		Str(l.NotificationIDKey, returnedID).
		Msg("Notification created successfully")
	return returnedID, nil
}

// Fetches notifications for a user
func GetNotificationByUserID(ctx context.Context,
	tx *sql.Tx,
	userID string,
	limit int,
) ([]dbModels.NotificationWithActor, error) {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"GetNotificationByUserID",
	).With().Str(l.UserIDKey, userID).Logger()

	query := `
  SELECT
    un.notification_id, un.recipient_user_id, un.type, un.actor_user_id,
    un.message, un.is_read, un.link, un.created_at, un.friendship_id,
    u.username AS actor_username,
    u.display_name AS actor_display_name,
    u.avatar_url AS actor_avatar_url
  FROM user_notifications un
  JOIN users u ON un.actor_user_id = u.user_id
  WHERE un.recipient_user_id = $1
  ORDER BY un.created_at DESC
  LIMIT $2;
  `
	rows, err := querier.QueryContext(ctx, query, userID, limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query notifications")
		return nil, fmt.Errorf("error querying notifications for user %s: %w", userID, err)
	}
	defer rows.Close()

	var notifications []dbModels.NotificationWithActor
	for rows.Next() {
		var n dbModels.NotificationWithActor
		if err := rows.Scan(
			&n.NotificationID, &n.RecipientUserID, &n.Type, &n.ActorUserID,
			&n.Message, &n.IsRead, &n.Link, &n.CreatedAt, &n.FriendshipID,
			&n.ActorUsername, &n.ActorDisplayName, &n.ActorAvatarURL,
		); err != nil {
			logger.Error().Err(err).Msg("Failed to scan notification row")
			return nil, fmt.Errorf("error scanning notification for user %s: %w", userID, err)
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// Marks a single notification as read or unread
func UpdateNotificationReadStatus(
	ctx context.Context,
	tx *sql.Tx,
	notificationID,
	userID string,
	readStatus bool,
) error {
	querier := GetQuerier(tx)
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		notificationComponent,
		"UpdateNotificationReadStatus",
	).With().
		Str(l.NotificationIDKey, notificationID).
		Str(l.UserIDKey, userID).
		Bool(l.NotificationReadKey, readStatus).
		Logger()

	query := `
  UPDATE user_notifications 
  SET is_read = $1 
  WHERE notification_id = $2 AND recipient_user_id = $3;
  `

	var err error
	if readStatus {
		_, err = querier.ExecContext(ctx, query, "TRUE", notificationID, userID)
	} else {
		_, err = querier.ExecContext(ctx, query, "FALSE", notificationID, userID)
	}
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update notification status")
		return fmt.Errorf("error updating notification status: %w", err)
	}
	return nil
}
