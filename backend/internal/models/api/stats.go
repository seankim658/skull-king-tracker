package models

// Represents a user's calculated statistics for display on a profile
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
	AwardType  string `json:"award_type"`
	AwardTitle string `json:"award_title"`
	Count      int    `json:"count"`
}

type UserAwardsStatsResponse []UserAwardStat
