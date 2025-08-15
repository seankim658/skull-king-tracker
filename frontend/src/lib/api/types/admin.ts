import type { Pagination } from "./api";

export interface UpdateReportStatusPayload {
  status: "payload" | "dismissed";
}

export interface SendNotificationPayload {
  message: string;
  user_ids: string[];
  is_broadcast: boolean;
}

export interface ReportUserPayload {
  reason: string;
}

export interface BanUserRequest {
  reason: string;
}

// --- Response Payloads ---

// Represents a single user in the admin users table
export interface AdminUserView {
  user_id: string;
  username: string;
  email?: string | null;
  display_name?: string | null;
  avatar_url?: string | null;
  avatar_source?: string | null;
  stats_privacy: string;
  role: string;
  is_banned: boolean;
  ban_reason?: string | null;
  created_at: string;
  updated_at: string;
  last_login_at?: string | null;
}

// Paginated response for a list of admin users
export interface PaginatedAdminUsersResponse {
  users: AdminUserView[];
  pagination: Pagination;
}
