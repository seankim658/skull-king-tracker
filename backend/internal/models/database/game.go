package models

import (
	"database/sql"
	"time"
)

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

	SessionName     sql.NullString `db:"session_name"`
	ScorekeeperName sql.NullString `db:"scorekeeper_name"`
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

// Helper struct for a game player's display details
type GamePlayerDetails struct {
	GamePlayerID  string         `db:"game_player_id"`
	GameID        string         `db:"game_id"`
	UserID        sql.NullString `db:"user_id"`
	GuestPlayerID sql.NullString `db:"guest_player_id"`
	DisplayName   string         `db:"display_name"`
	Username      sql.NullString `db:"username"`
	AvatarURL     sql.NullString `db:"avatar_url"`
	SeatingOrder  int            `db:"seating_order"`
	FinalScore    int            `db:"final_score"`
}

// Maps to the `guest_players` table
type GuestPlayer struct {
	GuestPlayerID string    `db:"guest_player_id"`
	DisplayName   string    `db:"display_name"`
	CreatedAt     time.Time `db:"created_at"`
}

// Helper struct for a game with winner information
type GameWithWinner struct {
	GameID              string         `db:"game_id"`
	Status              string         `db:"status"`
	CreatedAt           time.Time      `db:"created_at"`
	CompletedAt         sql.NullTime   `db:"completed_at"`
	WinningPlayer       sql.NullString `db:"winning_player"`
	IsViewerScorekeeper bool           `db:"is_viewer_scorekeeper"`
	ScorekeeperName     sql.NullString `db:"scorekeeper_name"`
}

// Contains the full details for a player's score in a round
type PlayerRoundScoreDetails struct {
	PlayerRoundScoreID string        `db:"player_round_score_id"`
	RoundID            string        `db:"round_id"`
	GamePlayerID       string        `db:"game_player_id"`
	BidAmount          sql.NullInt32 `db:"bid_amount"`
	TricksTaken        sql.NullInt32 `db:"tricks_taken"`
	RoundScore         int           `db:"round_score"`
	BonusPointsApplied int           `db:"bonus_points_applied"`
}

type Round struct {
	RoundID            string    `db:"round_id"`
	GameID             string    `db:"game_id"`
	RoundNumber        int       `db:"round_number"`
	DealerGamePlayerID string    `db:"dealer_game_player_id"`
	Status             string    `db:"status"`
	IsTiebreakerRound  bool      `db:"is_tiebreaker_round"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// Composite struct to hold all data for a scorecard
type FullScorecardData struct {
	Game    Game
	Players []GamePlayerDetails
	Rounds  []Round
	Scores  []PlayerRoundScoreDetails
}

type PlayerBidData struct {
	GamePlayerID string
	BidAmount    int
}

type PlayerScoreData struct {
	GamePlayerID string
	TricksTaken  int
	BonusPoints  int
}

type UserGameHistoryRow struct {
	GameID            string
	SessionName       sql.NullString
	GameDate          time.Time
	FinishingPosition sql.NullInt32
	TotalPoints       sql.NullInt32
	RoundsHit         sql.NullInt32
	ZeroDifferential  sql.NullInt32
	TotalPlayers      sql.NullInt32
	TotalAsterisks    sql.NullInt32
	ScorekeeperName   sql.NullString
}

type PlayerGameAsterisk struct {
	PlayerGameAsteriskID string         `db:"player_game_asterisk_id"`
	GamePlayerID         string         `db:"game_player_id"`
	GameID               string         `db:"game_id"`
	Reason               sql.NullString `db:"reason"`
	CreatedAt            time.Time      `db:"created_at"`
}
