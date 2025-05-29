package models

import "time"

type ActiveSessionResponse struct {
	SessionID     string     `json:"session_id"`
	SessionName   *string    `json:"session_name,omitempty"`
	Status        string     `json:"status"`
	HasActiveGame bool       `json:"has_active_game"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
