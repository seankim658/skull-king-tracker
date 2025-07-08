package models

import "time"

// --- Response Payloads ---

// Represents a single session in the user's list of active sessions
type ActiveSessionResponse struct {
	SessionID      string     `json:"session_id"`
	SessionName    *string    `json:"session_name,omitempty"`
	Status         string     `json:"status"`
	HasActiveGame  bool       `json:"has_active_game"`
	HasPendingGame bool       `json:"has_pending_game"`
	CreatorName    *string    `json:"creator_name,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// Response for a single session's detail page
type SessionDetailResponse struct {
	SessionID   string             `json:"session_id"`
	SessionName *string            `json:"session_name,omitempty"`
	Status      string             `json:"status"`
	Games       []SessionGame      `json:"games"`
	UserSummary SessionUserSummary `json:"user_summary"`
}

// Paginated response for a user's session history
type PaginatedSessionHistoryResponse struct {
	Sessions   []SessionHistoryItem `json:"sessions"`
	Pagination Pagination           `json:"pagination"`
}

// --- Component Structs ---

// Represents a single game within the session detail view
type SessionGame struct {
	GameID          string    `json:"game_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *string   `json:"completed_at,omitempty"`
	WinningPlayer   *string   `json:"winning_player,omitempty"`
	IsScorekeeper   bool      `json:"is_scorekeeper"`
	ScorekeeperName *string   `json:"scorekeeper_name,omitempty"`
}

// A summary of the viewing user's performance in a session
type SessionUserSummary struct {
	TotalGames int `json:"total_games"`
	Wins       int `json:"wins"`
}

// A single entry in the session history list
type SessionHistoryItem struct {
	SessionID                string    `json:"session_id"`
	SessionName              *string   `json:"session_name,omitempty"`
	DateCompleted            time.Time `json:"date_completed"`
	NumberOfGames            int       `json:"number_of_games"`
	YourWins                 int       `json:"your_wins"`
	WinPercentage            float64   `json:"win_percentage"`
	AverageFinishingPosition float64   `json:"average_finishing_position"`
	SessionCreator           string    `json:"session_creator"`
}
