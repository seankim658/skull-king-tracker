package models

type UserStats struct {
	TotalGamesPlayed int     `json:"total_games_played"`
	TotalWins        int     `json:"total_wins"`
	WinPercentage    float64 `json:"win_percentage"`
}
