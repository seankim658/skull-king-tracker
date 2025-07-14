export interface UserDetailedStats {
  total_games_played: number;
  total_wins: number;
  win_percentage: number;
  top_3_finishes: number;
  average_finishing_position: number;
  total_points: number;
  hit_percentage: number;
  total_zero_bids_made: number;
  zero_bid_success_rate: number;
}

export interface SiteSummaryStatsResponse {
  total_players: number;
  sessions_this_month: number;
  games_this_month: number;
  new_users_this_month: number;
}

export interface UserAwardStat {
  award_type: string;
  award_title: string;
  count: number;
  percentile: number;
}

export type UserAwardsStatsResponse = UserAwardStat[];

export interface GlobalLeaderboardItem {
  rank: number;
  user_id: string;
  player_name: string;
  games_played: number;
  wins: number;
  total_points: number;
  average_points: number;
  average_finish_pos: number;
}

export type GlobalLeaderboardResponse = GlobalLeaderboardItem[];
