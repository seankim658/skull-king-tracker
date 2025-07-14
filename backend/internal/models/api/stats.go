package models

// Represents a user's basic statistics
type UserStats struct {
	TotalGamesPlayed int     `json:"total_games_played"`
	TotalWins        int     `json:"total_wins"`
	Top3Finishes     int     `json:"top_3_finishes"`
	WinPercentage    float64 `json:"win_percentage"`
}

// Response for site-wide summary statistics
type SiteSummaryStatsResponse struct {
	TotalPlayers      int `json:"total_players"`
	SessionsThisMonth int `json:"sessions_this_month"`
	GamesThisMonth    int `json:"games_this_month"`
	NewUsersThisMonth int `json:"new_users_this_month"`
}

// Represents a single award's statistics
type UserAwardStat struct {
	AwardType  string  `json:"award_type"`
	AwardTitle string  `json:"award_title"`
	Count      int     `json:"count"`
	Percentile float64 `json:"percentile"`
}

type UserAwardsStatsResponse []UserAwardStat

type GlobalLeaderboardItem struct {
	Rank             int64   `json:"rank"`
	UserID           string  `json:"user_id"`
	PlayerName       string  `json:"player_name"`
	GamesPlayed      int     `json:"games_played"`
	Wins             int     `json:"wins"`
	TotalPoints      int     `json:"total_points"`
	AveragePoints    float64 `json:"average_points"`
	AverageFinishPos float64 `json:"average_finish_pos"`
}

type GlobalLeaderboardresponse []GlobalLeaderboardItem

// Represents a user's calculated statistics for display on a profile
type UserDetailedStats struct {
	TotalGamesPlayed         int     `json:"total_games_played"`
	TotalWins                int     `json:"total_wins"`
	WinPercentage            float64 `json:"win_percentage"`
	Top3Finishes             int     `json:"top_3_finishes"`
	AverageFinishingPosition float64 `json:"average_finishing_position"`
	TotalPoints              int     `json:"total_points"`
	HitPercentage            float64 `json:"hit_percentage"`
	TotalZeroBidsMade        int     `json:"total_zero_bids_made"`
	ZeroBidSuccessRate       float64 `json:"zero_bid_success_rate"`
}
