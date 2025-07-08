package models

import "database/sql"

// Holds a user's basic game statistics
type ProfileStats struct {
	TotalGamesPlayed int
	TotalWins        int
	Top3Finishes     int
}

// Holds basic site-wide summary statistics
type SiteWideSummaryStats struct {
	TotalPlayers      int
	SessionsThisMonth int
	GamesThisMonth    int
	NewUsersThisMonth int
}

// Holds the calculated statistics for a single player in a game
type GameSummaryPlayerStats struct {
	GamePlayerID          string
	DisplayName           string
	FinalScore            int
	FinishingPosition     sql.NullInt32
	RoundsHit             int
	RoundsMissed          int
	ZeroBidsHit           int
	TotalBonus            int
	TotalTricksTaken      int
	BidStdDev             sql.NullFloat64
	PointsFromCorrectBids int
	TricksFromCorrectBids int
	AvgBid                sql.NullFloat64
}
