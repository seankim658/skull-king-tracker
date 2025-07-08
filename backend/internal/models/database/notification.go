package models

import (
	"database/sql"
	"time"
)

// Maps to the `user_notifications` table
type Notification struct {
	NotificationID  string         `db:"notification_id"`
	RecipientUserID string         `db:"recipient_user_id"`
	Type            string         `db:"type"`
	ActorUserID     string         `db:"actor_user_id"`
	Message         string         `db:"message"`
	IsRead          bool           `db:"is_read"`
	Link            sql.NullString `db:"link"`
	FriendshipID    sql.NullString `db:"friendship_id"`
	CreatedAt       time.Time      `db:"created_at"`
}

// Composite struct that includes the notification and details about the actor
type NotificationWithActor struct {
	Notification
	ActorUsername    string         `db:"actor_username"`
	ActorDisplayName sql.NullString `db:"actor_display_name"`
	ActorAvatarURL   sql.NullString `db:"actor_avatar_rul"`
}
