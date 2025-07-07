package models

import "time"

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

type SessionGame struct {
	GameID          string    `json:"game_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *string   `json:"completed_at,omitempty"`
	WinningPlayer   *string   `json:"winning_player,omitempty"`
	IsScorekeeper   bool      `json:"is_scorekeeper"`
	ScorekeeperName *string   `json:"scorekeeper_name,omitempty"`
}

type SessionUserSummary struct {
	TotalGames int `json:"total_games"`
	Wins       int `json:"wins"`
}

type SessionDetailResponse struct {
	SessionID   string             `json:"session_id"`
	SessionName *string            `json:"session_name,omitempty"`
	Status      string             `json:"status"`
	Games       []SessionGame      `json:"games"`
	UserSummary SessionUserSummary `json:"user_summary"`
}

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

type PaginatedSessionHistoryResponse struct {
	Sessions   []SessionHistoryItem `json:"sessions"`
	Pagination Pagination           `json:"pagination"`
}
