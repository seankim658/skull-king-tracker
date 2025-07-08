import type { Pagination } from "./api";

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
  current_scorekeeper_user_id?: string;
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
  display_name: string;
  username?: string;
  avatar_url?: string;
  seating_order: number;
  final_score: number;
}

/**
 * Payload for updating game setings.
 */
export interface UpdateGameSettingsPayload {
  scorekeeper_user_id: string;
  ordered_player_ids: string[];
  starting_dealer_game_player_id: string;
}

/**
 * Holds the score details for a single player in a single round.
 */
export interface PlayerRoundData {
  game_player_id: string;
  bid_amount: number | null;
  tricks_taken: number | null;
  round_score: number | null;
  bonus_points: number | null;
}

/**
 *  Contains all player data for a single round of the scorecard.
 */
export interface RoundScorecard {
  round_number: number;
  status: "bidding" | "playing" | "completed";
  player_scores: PlayerRoundData[];
  dealer_game_player_id: string;
}

/**
 * The main object for fetching and rendering the entire scorecard.
 */
export interface ScorecardResponse {
  game_id: string;
  game_status: string;
  players: GamePlayerResponse[];
  rounds: RoundScorecard[];
  current_round: number;
  session_name?: string | null;
  current_scorekeeper_user_id: string;
  scorekeeper_name: string;
  asterisks: PlayerGameAsterisk[];
}

/**
 * Defines the structure for a single player's bid submission.
 */
export interface PlayerBid {
  game_player_id: string;
  bid_amount: number;
}

/**
 * Payload for submitting all bids for a round.
 */
export interface SubmitBidsPayload {
  bids: PlayerBid[];
}

/**
 * Defines the structure for a single player's trick submission.
 */
export interface PlayerTricks {
  game_player_id: string;
  tricks_taken: number;
  bonus_points: number;
}

/**
 * Payload for submitting all tricks taken for a round.
 */
export interface SubmitTricksPayload {
  tricks: PlayerTricks[];
}

/**
 * Represents a single player with an active game card for the dashboard.
 */
export interface ActiveGamePlayer {
  display_name: string;
  avatar_url?: string | null;
}

/**
 * Represents a single active game card for the dashboard list.
 */
export interface ActiveGameResponse {
  game_id: string;
  session_name?: string | null;
  scorekeeper_name: string;
  is_scorekeeper: boolean;
  created_at: string;
  current_round: number;
  players: ActiveGamePlayer[];
}

export interface GameHistoryItem {
  game_id: string;
  session_name?: string | null;
  game_date: string;
  finishing_position: number;
  total_points: number;
  rounds_hit: number;
  zero_differential: number;
  total_players: number;
  total_asterisks: number;
  scorekeeper_name: string;
}

export interface PaginatedGameHistoryResponse {
  games: GameHistoryItem[];
  pagination: Pagination;
}

/**
 * Payload for adding an asterisk to a player.
 */
export interface AddAsteriskPayload {
  reason: string;
}

/**
 * Represents a single player asterisk.
 */
export interface PlayerGameAsterisk {
  player_game_asterisk_id: string;
  game_player_id: string;
  reason?: string | null;
  created_at: string;
}
