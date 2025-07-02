import { client } from "../client";
import type { ApiResponse, SiteSummaryStatsResponse } from "../types";

export const statsAPI = {
  /**
   * Fetches the site-wide summary statistics.
   */
  getSiteSummaryStats: (): Promise<ApiResponse<SiteSummaryStatsResponse>> =>
    client<SiteSummaryStatsResponse>("/stats/summary", {
      method: "GET",
    }),
};
