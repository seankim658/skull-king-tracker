import type { ApiResponse } from "./types";

/**
 * Base URL for all API requests.
 */
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

/**
 * Core function to make API requests with proper error handling.
 *
 * @template T - The expected type of the response data.
 * @param {string} endpoint - The API endpoint to call.
 * @param {RequestOptions} options - Request options including HTTP method, body, and handlers.
 * @returns Promise that resolves to the full API response.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function request<T = any>(
  endpoint: string,
  { body, ...customConfig }: RequestInit = {},
): Promise<ApiResponse<T>> {
  const headers: HeadersInit = {};

  if (body && !(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const config: RequestInit = {
    ...customConfig,
    headers: {
      ...headers,
      ...customConfig.headers,
    },
  };

  if (body) {
    config.body = body;
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

    if (response.status === 204) {
      return { success: true };
    }

    const data = await response.json();

    if (!response.ok) {
      return Promise.reject(
        new Error(
          data.message ||
            data.error ||
            `API Error: ${response.statusText} (${response.status})`,
        ),
      );
    }

    return data as ApiResponse<T>;
  } catch (e) {
    console.error("API Client Error:", e);
    return Promise.reject(e);
  }
}

/**
 * Public client function used by all API services to make requests.
 *
 * @template T - The expected type of the response data.
 * @param {string} endpoint - The API endpoint t ocall.
 * @param {RequestOptions} options - Request options.
 * @returns Promise that resolves to the API response data.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function client<T = any>(endpoint: string, options: RequestInit = {}) {
  return request<T>(endpoint, options);
}
