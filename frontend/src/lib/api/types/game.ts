/**
 * Payload for creating a new game.
 */
export interface CreateGamePayload {
  session_id?: string;
  session_name?: string;
}

/**
 * Response for a created game.
 */
export interface GameResponse {
  game_id: string;
  session_id?: string;
  status: string;
  created_at: string;
  created_by_user_id: string;
}

/**
 * Payload for adding a player to a game.
 */
export interface AddPlayerToGamePayload {
  user_id?: string;
  guest_name?: string;
  seating_order: number;
}

/**
 * Response for an added game player.
 */
export interface GamePlayerResponse {
  game_player_id: string;
  game_id: string;
  user_id?: string;
  guest_player_id?: string;
  display_naem: string;
  seating_order: number;
  final_score: number;
}
