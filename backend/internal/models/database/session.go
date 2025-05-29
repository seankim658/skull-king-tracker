package models

import (
	"database/sql"
	"time"
)

// Mpas to the `game_sessions` table
type GameSession struct {
	SessionID       string         `db:"session_id"`
	SessionName     sql.NullString `db:"session_name"`
	CreatedByUserID sql.NullString `db:"created_by_user_id"`
	Status          string         `db:"status"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	CompletedAt     sql.NullTime   `db:"completed_at"`
}
