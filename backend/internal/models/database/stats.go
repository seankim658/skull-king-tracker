package models

import "database/sql"

// Holds a user's basic game statistics for their profile or a specific session
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
	RoundScoreVariance    sql.NullFloat64
	PointsLastThreeRounds int
	BiggestBust           int
	FailedZeroBids        int
	TotalAsterisks        int
}

// Summary of a single award type for a user
type UserAwardStat struct {
	AwardType     string  `db:"award_type"`
	AwardCount    int     `db:"award_count"`
	PercentilRank float64 `db:"percentile_rank"`
}

type GlobalLeaderBoardRow struct {
	Rank             int64
	UserID           string
	DisplayName      string
	GamesPlayed      int
	Wins             int
	TotalPoints      int
	AveragePoints    float64
	AverageFinishPos float64
}

type UserDetailedStats struct {
	TotalGamesPlayed         int
	TotalWins                int
	Top3Finishes             int
	AverageFinishingPosition float64
	TotalPoints              int
	TotalRoundsPlayed        int
	TotalBidsHit             int
	TotalZeroBidsMade        int
	SuccessfulZeroBids       int
}
