/**
 * Response for an active session.
 */
export interface ActiveSessionResponse {
  session_id: string;
  session_name?: string;
  status: string;
  has_active_game: boolean;
  has_pending_game: boolean;
  creator_name?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string | null;
}

/**
 * Represents a single game within a session for the detail modal.
 */
export interface SessionGame {
  game_id: string;
  status: string;
  created_at: string;
  completed_at?: string | null;
  winning_player?: string | null;
  is_scorekeeper: boolean;
  scorekeeper_name?: string | null;
}

/**
 * Represents the user's summary stats for a session.
 */
export interface SessionUserSummary {
  total_games: number;
  wins: number;
}

/**
 * Response for session details.
 */
export interface SessionDetailResponse {
  session_id: string;
  session_name?: string;
  status: string;
  games: SessionGame[];
  user_summary: SessionUserSummary;
}
