package models

import "time"

// Request to create a new game
type CreateGameRequest struct {
	SessionID   *string `json:"session_id,omitempty"`
	SessionName *string `json:"session_name,omitempty"`
}

// Request to add a player to a game
type AddPlayerToGameRequest struct {
	UserID       *string `json:"user_id,omitempty"`
	GuestName    *string `json:"guest_name,omitempty"`
	SeatingOrder int     `json:"seating_order" validate:"required,gt=0"`
}

// Response for a created game
type GameResponse struct {
	GameID          string    `json:"game_id"`
	SessionID       *string   `json:"session_id,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedByUserID string    `json:"created_by_user_id"`
}

type GamePlayerResponse struct {
	GamePlayerID  string  `json:"game_player_id"`
	GameID        string  `json:"game_id"`
	UserID        *string `json:"user_id,omitempty"`
	GuestPlayerID *string `json:"guest_player_id,omitempty"`
	DisplayName   string  `json:"display_name"`
	SeatingOrder  int     `json:"seating_order"`
	FinalScore    int     `json:"final_score"`
}
