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
 * @returns Promise that resolves to the API response data.
 */
async function request<T>(
  endpoint: string,
  { body, ...customConfig }: RequestInit = {},
): Promise<T> {
  const headers: HeadersInit = {};

  // Only set Content-Type if body is present and not FormData
  if (body && !(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  // Create the final request configuration
  const config: RequestInit = {
    ...customConfig,
    headers: {
      ...headers,
      ...customConfig.headers,
    },
  };

  // Add body to config if it's not a GET or HEAD request
  if (body) {
    config.body = body;
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch (e) {
        console.error("Error extracting response error:", e);
        errorData = {
          error: `API Error: ${response.statusText} (${response.status})`,
        };
      }
      return Promise.reject(
        new Error(
          errorData?.error ||
            errorData?.message ||
            `API request failed with status ${response.status}`,
        ),
      );
    }

    // Handle responses with no content
    if (response.status === 204) {
      return {} as T;
    }

    // For other successful responses, parse JSON
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      return (await response.json()) as T;
    }

    return (await response.text()) as unknown as T;
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
export function client<T>(endpoint: string, options: RequestInit = {}) {
  return request<T>(endpoint, options);
}
