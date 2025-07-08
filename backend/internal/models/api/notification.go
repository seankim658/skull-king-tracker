package models

import "time"

const (
	NotificationTypeFriendRequest = "friend_request"
)

// Represents the user who performed the action that triggered a notification
type NotificationActor struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// Represents a single notification sent to the client
type Notification struct {
	NotificationID string            `json:"notification_id"`
	Type           string            `json:"type"`
	Actor          NotificationActor `json:"actor"`
	Message        string            `json:"message"`
	IsRead         bool              `json:"is_read"`
	FriendshipID   *string           `json:"friendship_id,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}
