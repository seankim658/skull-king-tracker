package models

import (
	"database/sql"
	"time"
)

// Maps to the `games` table
type Game struct {
	GameID                       string         `db:"game_id"`
	SessionID                    sql.NullString `db:"session_id"`
	CreatedByUserID              string         `db:"created_by_user_id"`
	CurrentScorekeeperUserID     sql.NullString `db:"current_scorekeeper_user_id"`
	Status                       string         `db:"status"`
	StartingDealerGamePlayerID   sql.NullString `db:"starting_dealer_game_player_id"`
	PlayerSeatingOrderRandomized bool           `db:"player_seating_order_randomized"`
	CreatedAt                    time.Time      `db:"created_at"`
	UpdatedAt                    time.Time      `db:"updated_at"`
	CompletedAt                  sql.NullTime   `db:"completed_at"`
}

// Maps to the `game_players` table
type GamePlayer struct {
	GamePlayerID      string         `db:"game_player_id"`
	GameID            string         `db:"game_id"`
	UserID            sql.NullString `db:"user_id"`
	GuestPlayerID     sql.NullString `db:"guest_player_id"`
	SeatingOrder      int            `db:"seating_order"`
	FinalScore        int            `db:"final_score"`
	FinishingPosition sql.NullInt32  `db:"finishing_position"`
}

// Maps to the `guest_players` table
type GuestPlayer struct {
	GuestPlayerID string    `db:"guest_player_id"`
	DisplayName   string    `db:"display_name"`
	CreatedAt     time.Time `db:"created_at"`
}
