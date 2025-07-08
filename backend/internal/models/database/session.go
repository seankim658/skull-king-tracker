package models

import (
	"database/sql"
	"time"
)

// --- Core Entities ---

// Maps to the `game_sessions` table
type GameSession struct {
	SessionID       string         `db:"session_id"`
	SessionName     sql.NullString `db:"session_name"`
	CreatedByUserID sql.NullString `db:"created_by_user_id"`
	Status          string         `db:"status"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	CompletedAt     sql.NullTime   `db:"completed_at"`
}

// --- Composite & Helper Structs ---

// Helper struct to include session details along with game flow activity
type GameSessionWithActivity struct {
	GameSession
	HasActiveGame  bool           `db:"has_active_game"`
	HasPendingGame bool           `db:"has_pending_game"`
	CreatorName    sql.NullString `db:"creator_name"`
}

// Represents a single row returned for a user's paginated session history
type UserSessionHistoryRow struct {
	SessionID              string
	SessionName            sql.NullString
	DateCompleted          time.Time
	NumberOfGames          int
	YourWins               int
	TotalFinishingPosition int
	SessionCreator         sql.NullString
}

// Composite struct holding all data needed for the session detail view
type SessionDetailData struct {
	Session   GameSession
	Games     []GameWithWinner
	UserStats ProfileStats
}
