import { client } from "../client";
import type {
  UserAwardsStatsResponse,
  ApiResponse,
  SiteSummaryStatsResponse,
} from "../types";

export const statsAPI = {
  /**
   * Fetches the site-wide summary statistics.
   */
  getSiteSummaryStats: (): Promise<ApiResponse<SiteSummaryStatsResponse>> =>
    client<SiteSummaryStatsResponse>("/stats/summary", {
      method: "GET",
    }),

  /**
   * Fetches the awards summary for a specific user.
   */
  getUserAwardsStats: (
    userId: string,
  ): Promise<ApiResponse<UserAwardsStatsResponse>> =>
    client<UserAwardsStatsResponse>(`/users/${userId}/stats/awards`, {
      method: "GET",
    }),
};
