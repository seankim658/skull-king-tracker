package models

type SSEEvent struct {
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

type SSEDeletedNotificationPayload struct {
	NotificationID string `json:"notification_id"`
}
