import { client } from "../client";
import type {
  ApiResponse,
  ActiveSessionResponse,
  SessionDetailResponse,
} from "../types";

export const sessionAPI = {
  /**
   * Fetches active game sessions for the current user.
   */
  getActiveSessionsForUser: (): Promise<ApiResponse<ActiveSessionResponse[]>> =>
    client<ActiveSessionResponse[]>("/sessions/active", {
      method: "GET",
    }),

  /**
   * Fetches the detailed information for a single session.
   * @param sessionId - The ID of the session to fetch details for
   */
  getSessionDetails: (
    sessionId: string,
  ): Promise<ApiResponse<SessionDetailResponse>> =>
    client<SessionDetailResponse>(`/sessions/${sessionId}`, {
      method: "GET",
    }),

  /**
   * Marks a game session as 'completed'.
   * @param sessionid - The ID of the session to complete
   */
  completeSession: (sessionId: string): Promise<ApiResponse<null>> =>
    client<null>(`/sessions/${sessionId}/complete`, {
      method: "PUT",
    }),
};
