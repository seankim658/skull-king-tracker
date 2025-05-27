import { client } from "../client";
import type { ApiResponse, AuthenticatedUserResponse } from "../types";

/**
 * Represents the expected response from the logout endpoint.
 */
export interface LogoutResponse {
  message: string;
}

/**
 * Authentication API service methods.
 */
export const authAPI = {
  /**
   * Fetches the current authenticated user's details from the backend.
   * @returns Promise resolving to the user's data if authenticated
   */
  getCurrentUser: (): Promise<ApiResponse<AuthenticatedUserResponse>> =>
    client<ApiResponse<AuthenticatedUserResponse>>("/auth/me", {
      method: "GET",
    }),

  /**
   * Logs out the current user.
   * @returns Promise resolving to a success message.
   */
  logout: (): Promise<ApiResponse<LogoutResponse>> =>
    client<ApiResponse<LogoutResponse>>("/auth/logout", {
      method: "POST",
    }),
};
