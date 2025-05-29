/**
 * Response for an active session.
 */
export interface ActiveSessionResponse {
  session_id: string;
  session_name?: string;
  status: string;
  has_active_game: boolean;
  created_at: string;
  updated_at: string;
  completed_at?: string | null;
}
