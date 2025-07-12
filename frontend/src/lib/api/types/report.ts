import type { Pagination } from "./api";

export interface UserReport {
  report_id: string;
  reporter_user_id: string;
  reporter_name: string;
  reported_user_id: string;
  reported_name: string;
  reason: string;
  status: "pending" | "resolved" | "dismissed";
  created_at: string;
}

export interface PaginatedReportsResponse {
  reports: UserReport[];
  pagination: Pagination;
}
