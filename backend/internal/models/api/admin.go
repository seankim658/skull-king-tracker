package models

import "time"

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

// --- Response Payloads ---

// Represents a single user in the admin users table
type AdminUserViewResponse struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Email        *string   `json:"email,omitempty"`
	DisplayName  *string   `json:"display_name,omitempty"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	AvatarSource *string   `json:"avatar_source,omitempty"`
	StatsPrivacy string    `json:"stats_privacy"`
	Role         string    `json:"role"`
	IsBanned     bool      `json:"is_banned"`
	BanReason    *string   `json:"ban_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  *string   `json:"last_login_at,omitempty"`
}

// Paginated response for a list of admin users
type PaginatedUsersResponse struct {
	Users      []AdminUserViewResponse `json:"users"`
	Pagination Pagination              `json:"pagination"`
}
