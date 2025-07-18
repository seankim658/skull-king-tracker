import { client } from "../client";
import type {
  ApiResponse,
  PaginatedAlertsResponse,
  SiteAlert,
  SiteAlertPayload,
} from "../types";
import type { SortingState } from "@tanstack/react-table";

export interface AlertFilters {
  page: number;
  pageSize: number;
  sorting: SortingState;
}

export const alertAPI = {
  /**
   * Fetches currently active alerts for public display.
   */
  getActiveAlerts: (): Promise<ApiResponse<SiteAlert[]>> =>
    client<SiteAlert[]>("/alerts/active", { method: "GET" }),

  /**
   * Fetches a paginated list of all alerts.
   */
  getAdminAlerts: ({
    page,
    pageSize,
    sorting,
  }: AlertFilters): Promise<ApiResponse<PaginatedAlertsResponse>> => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (sorting.length > 0) {
      params.append("sort_by", sorting[0].id);
      params.append("sort_order", sorting[0].desc ? "desc" : "asc");
    }
    return client<PaginatedAlertsResponse>(
      `/admin/alerts?${params.toString()}`,
    );
  },

  /**
   * Creates a new site alert.
   */
  createAlert: (payload: SiteAlertPayload): Promise<ApiResponse<null>> =>
    client<null>("/admin/alerts", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  /**
   * Updates an existing site alert.
   */
  updateAlert: (
    alertId: string,
    payload: SiteAlertPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/admin/alerts/${alertId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),

  /**
   * Deletes a site alert.
   */
  deleteAlert: (alertId: string): Promise<ApiResponse<null>> =>
    client<null>(`/admin/alerts/${alertId}`, {
      method: "DELETE",
    }),
};
