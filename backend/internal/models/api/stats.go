package models

type UserStats struct {
	TotalGamesPlayed int     `json:"total_games_played"`
	TotalWins        int     `json:"total_wins"`
	WinPercentage    float64 `json:"win_percentage"`
}

type SiteSummaryStatsResponse struct {
	TotalPlayers      int `json:"total_players"`
	SessionsThisMonth int `json:"sessions_this_month"`
	GamesThisMonth    int `json:"games_this_month"`
	NewUsersThisMonth int `json:"new_users_this_month"`
}
