export interface UserStats {
  total_games_played: number;
  total_wins: number;
  top_3_finishes: number;
  win_percentage: number;
}

export interface SiteSummaryStatsResponse {
  total_players: number;
  sessions_this_month: number;
  games_this_month: number;
  new_users_this_month: number;
}
