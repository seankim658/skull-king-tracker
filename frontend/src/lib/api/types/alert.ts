import type { Pagination } from "./api";

/**
 * Represents a single site alert.
 */
export interface SiteAlert {
  alert_id: string;
  title: string;
  body: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
  creator_name: string;
  created_at: string;
  updated_at: string;
}

/**
 * Payload for creating or updating a site alert.
 */
export interface SiteAlertPayload {
  title: string;
  body: string;
  start_time: Date | string;
  end_time: Date | string;
  is_active: boolean;
}

/**
 * Paginated response for a list of site alerts.
 */
export interface PaginatedAlertsResponse {
  alerts: SiteAlert[];
  pagination: Pagination;
}
