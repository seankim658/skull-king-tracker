import { client } from "../client";
import type {
  ApiResponse,
  AuthenticatedUserResponse,
  UpdateUserThemePayload,
  UpdateUserProfilePayload,
  LinkedAccount,
  UserProfileResponse,
  UserSearchResponse,
} from "../types";

export const userAPI = {
  /**
   * Updates the user's theme settings.
   * @param payload - The new UI and color theme
   * @returns Promise resolving to the updated user data
   */
  updateThemeSettings: (
    payload: UpdateUserThemePayload,
  ): Promise<ApiResponse<AuthenticatedUserResponse>> =>
    client<ApiResponse<AuthenticatedUserResponse>>("/settings/theme", {
      method: "PUT",
      body: JSON.stringify(payload),
      headers: {
        "Content-Type": "application/json",
      },
    }),

  /**
   * Updates the user's profile information.
   * @param payload - The new profile data
   * @returns Promise resolving to the updated user data
   */
  updateUserProfile: (
    payload: UpdateUserProfilePayload,
  ): Promise<ApiResponse<AuthenticatedUserResponse>> =>
    client<ApiResponse<AuthenticatedUserResponse>>("/settings/profile", {
      method: "PUT",
      body: JSON.stringify(payload),
      headers: {
        "Content-Type": "application/json",
      },
    }),

  /**
   * Retrieves the list of OAuth accounts linked to the current user.
   * @returns Promise resolving to the list of linked accounts
   */
  getLinkedAccounts: (): Promise<ApiResponse<LinkedAccount[]>> =>
    client<ApiResponse<LinkedAccount[]>>("/settings/linked-accounts", {
      method: "GET",
    }),

  /**
   * Unlinks an OAuth provider account for the current user.
   * @param providerName - The name of the provider to unlink
   * @returns Promise resolving to an API response
   */
  unlinkAccount: (providerName: string): Promise<ApiResponse<null>> =>
    client<ApiResponse<null>>(`/settings/linked-accounts/${providerName}`, {
      method: "DELETE",
    }),

  /**
   * Retrieves a user's public profile information.
   * @param userId - The ID of the user whose profile to fetch
   * @returns Promise resolving to the user's profile data
   */
  getUserProfile: (userId: string): Promise<ApiResponse<UserProfileResponse>> =>
    client<ApiResponse<UserProfileResponse>>(`/users/${userId}/profile`, {
      method: "GET",
    }),

  /**
   * Searches for users by username or display name.
   * @param query - The search query string
   * @param limit - Optional limit for the number of results
   */
  searchUsers: (
    query: string,
    limit: number = 10,
  ): Promise<ApiResponse<UserSearchResponse>> => {
    const params = new URLSearchParams();
    params.append("q", query);
    params.append("limit", limit.toString());
    return client<ApiResponse<UserSearchResponse>>(
      `/users/search?${params.toString()}`,
      {
        method: "GET",
      },
    );
  },
};
