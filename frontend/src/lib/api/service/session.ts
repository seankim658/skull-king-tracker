import { client } from "../client";
import type {
  ApiResponse,
  ActiveSessionResponse,
  SessionDetailResponse,
  PaginatedSessionHistoryResponse,
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

  /**
   * Fetches a paginated history of a user's completed sessions.
   */
  getSessionHistory: (
    page: number,
    pageSize: number,
    sorting: { id: string; desc: boolean }[],
  ): Promise<ApiResponse<PaginatedSessionHistoryResponse>> => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (sorting.length > 0) {
      params.append("sort_by", sorting[0].id);
      params.append("sort_order", sorting[0].desc ? "desc" : "asc");
    }
    return client<PaginatedSessionHistoryResponse>(
      `/sessions/history?${params.toString()}`,
    );
  },
};
