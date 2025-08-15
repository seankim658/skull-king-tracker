import { client } from "../client";
import type {
  ApiResponse,
  PaginatedReportsResponse,
  UpdateReportStatusPayload,
  SendNotificationPayload,
  BanUserRequest,
  PaginatedAdminUsersResponse,
} from "../types";
import type { SortingState } from "@tanstack/react-table";

export interface ReportFilters {
  page: number;
  pageSize: number;
  status?: string;
  sorting: SortingState;
}

export interface UserFilters {
  page: number;
  pageSize: number;
}

export const adminAPI = {
  /**
   * Fetches user reports with optional filters and pagination.
   */
  getReports: ({
    page,
    pageSize,
    status,
    sorting,
  }: ReportFilters): Promise<ApiResponse<PaginatedReportsResponse>> => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    if (status) {
      params.append("status", status);
    }
    if (sorting.length > 0) {
      params.append("sort_by", sorting[0].id);
      params.append("sort_order", sorting[0].desc ? "desc" : "asc");
    }
    return client<PaginatedReportsResponse>(
      `/admin/reports?${params.toString()}`,
    );
  },

  /**
   * Updates the status of a report.
   */
  updateReportStatus: (
    reportId: string,
    payload: UpdateReportStatusPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/admin/reports/${reportId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),

  /**
   * Bans a user.
   */
  banUser: (
    userId: string,
    payload: BanUserRequest,
  ): Promise<ApiResponse<null>> =>
    client<null>(`/admin/users/${userId}/ban`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  /**
   * Unbans a user.
   */
  unbanUser: (userId: string): Promise<ApiResponse<null>> =>
    client<null>(`/admin/users/${userId}/unban`, { method: "POST" }),

  /**
   * Sends a notification from an admin. Can be a broadcast or targeted.
   */
  sendAdminNotification: (
    payload: SendNotificationPayload,
  ): Promise<ApiResponse<null>> =>
    client<null>("/admin/notifications", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  /**
   * Fetches all users for the admin panel.
   */
  getUsers: ({
    page,
    pageSize,
  }: UserFilters): Promise<ApiResponse<PaginatedAdminUsersResponse>> => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });
    return client<PaginatedAdminUsersResponse>(
      `/admin/users?${params.toString()}`,
    );
  },
};
