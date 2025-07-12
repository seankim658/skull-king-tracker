package models

// --- Request Payloads ---

// Payload to ban a user account
type BanUserRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// Payload to update a report status
type UpdateReportRequest struct {
	Status string `json:"status" validate:"required,oneof=resolved dismissed"`
}

// Payload to send a notification
type SendNotificationRequest struct {
	Message     string   `json:"message" validate:"required"`
	UserIDs     []string `json:"user_ids,omitempty"`
	IsBroadcast bool     `json:"is_broadcast"`
}
