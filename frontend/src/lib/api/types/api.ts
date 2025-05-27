/**
 * Standardized API response structure.
 * @tempate T - The type of the data payload in a successful response.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

/**
 * Standardized paginatino information.
 */
export interface Pagination {
  current_page: number;
  total_pages: number;
  page_size: number;
  total_count: number;
}
