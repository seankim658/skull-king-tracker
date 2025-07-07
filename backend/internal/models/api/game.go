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
	GameID                   string    `json:"game_id"`
	SessionID                *string   `json:"session_id,omitempty"`
	Status                   string    `json:"status"`
	CreatedAt                time.Time `json:"created_at"`
	CreatedByUserID          string    `json:"created_by_user_id"`
	CurrentScoreKeeperUserID *string   `json:"current_scorekeeper_user_id,omitempty"`
}

type GamePlayerResponse struct {
	GamePlayerID  string  `json:"game_player_id"`
	GameID        string  `json:"game_id"`
	UserID        *string `json:"user_id,omitempty"`
	GuestPlayerID *string `json:"guest_player_id,omitempty"`
	DisplayName   string  `json:"display_name"`
	Username      *string `json:"username,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	SeatingOrder  int     `json:"seating_order"`
	FinalScore    int     `json:"final_score"`
}

type UpdateGameSettingsRequest struct {
	ScorekeeperUserID          string   `json:"scorekeeper_user_id" validate:"required"`
	OrderedPlayerIDs           []string `json:"ordered_player_ids" validate:"required"`
	StartingDealerGamePlayerID string   `json:"starting_dealer_game_player_id" validate:"required"`
}

// Contains all player data for a single round
type PlayerRoundData struct {
	GamePlayerID string `json:"game_player_id"`
	BidAmount    *int   `json:"bid_amount,omitempty"`
	TricksTaken  *int   `json:"tricks_taken,omitempty"`
	RoundScore   *int   `json:"round_score,omitempty"`
	BonusPoints  *int   `json:"bonus_points,omitempty"`
}

// Main object for rendering the scorecard
type RoundScorecard struct {
	RoundNumber        int               `json:"round_number"`
	Status             string            `json:"status"`
	PlayerScores       []PlayerRoundData `json:"player_scores"`
	DealerGamePlayerID string            `json:"dealer_game_player_id"`
}

type ScorecardResponse struct {
	GameID                   string               `json:"game_id"`
	GameStatus               string               `json:"game_status"`
	Players                  []GamePlayerResponse `json:"players"`
	Rounds                   []RoundScorecard     `json:"rounds"`
	CurrentRound             int                  `json:"current_round"`
	SessionName              *string              `json:"session_name,omitempty"`
	CurrentScoreKeeperUserID string               `json:"current_scorekeeper_user_id"`
	ScorekeeperName          string               `json:"scorekeeper_name"`
}

type PlayerBid struct {
	GamePlayerID string `json:"game_player_id" validate:"required"`
	BidAmount    int    `json:"bid_amount" validate:"gte=0"`
}

// Payload for submitting all bids for a round
type SubmitBidsRequest struct {
	Bids []PlayerBid `json:"bids" validate:"required,min=1,dive"`
}

// Payload for submitting all tricks taken for a round
type SubmitTricksRequest struct {
	Tricks []PlayerTrickData `json:"tricks" validate:"required,min=1,dive"`
}

type PlayerTrickData struct {
	GamePlayerID string `json:"game_player_id" validate:"required"`
	TricksTaken  int    `json:"tricks_taken" validate:"gte=0"`
	BonusPoints  int    `json:"bonus_points" validate:"gte=0"`
}

// Represetns a single player within an active game card
type ActiveGamePlayer struct {
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// Represents a single active game for the dashboard list
type ActiveGameResponse struct {
	GameID          string             `json:"game_id"`
	SessionName     *string            `json:"session_name,omitempty"`
	ScorekeeperName string             `json:"scorekeeper_name"`
	IsScorekeeper   bool               `json:"is_scorekeeper"`
	CreatedAt       time.Time          `json:"created_at"`
	CurrentRound    int                `json:"current_round"`
	Players         []ActiveGamePlayer `json:"players"`
}

// Represents a single row in the user's game history table
type GameHistoryItem struct {
	GameID            string    `json:"game_id"`
	SessionName       *string   `json:"session_name,omitempty"`
	GameDate          time.Time `json:"game_date"`
	FinishingPosition int       `json:"finishing_position"`
	TotalPoints       int       `json:"total_points"`
	RoundsHit         int       `json:"rounds_hit"`
	ZeroDifferential  int       `json:"zero_differential"`
	TotalPlayers      int       `json:"total_players"`
	TotalAsterisks    int       `json:"total_asterisks"`
	ScorekeeperName   string    `json:"scorekeeper_name"`
}

// Represents the paginated response for the game history
type PaginatedGameHistoryResponse struct {
	Games      []GameHistoryItem `json:"games"`
	Pagination Pagination        `json:"pagination"`
}

type AddAsteriskRequest struct {
	Reason string `json:"reason"`
}

type PlayerGameAsterisk struct {
	PlayerGameAsteriskID string    `json:"player_game_asterisk_id"`
	GamePlayerID         string    `json:"game_player_id"`
	Reason               *string   `json:"reason,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}
