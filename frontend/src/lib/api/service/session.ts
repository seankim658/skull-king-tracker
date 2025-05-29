import { client } from "../client";
import type { ApiResponse, ActiveSessionResponse } from "../types";

export const sessionAPI = {
  /**
   * Fetches active game sessions for the current user.
   */
  getActiveSessionsForUser: (): Promise<ApiResponse<ActiveSessionResponse[]>> =>
    client<ApiResponse<ActiveSessionResponse[]>>("/sessions/active", {
      method: "GET",
    }),

  /**
   * Marks a game session as 'completed'.
   * @param sessionid - The ID of the session to complete
   */
  completeSession: (sessionId: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/sessions/${sessionId}/complete`, {
      method: "PUT",
    }),
};
