import type { UITheme, ColorTheme } from "@/lib/themes";

/**
 * Represents a user in the skull king application.
 */
export interface User {
  user_id: string;
  username: string;
  email: string | null;
  display_name: string | null;
  avatar_url: string | null;
  stats_privacy: "private" | "friends_only" | "public";
  ui_theme: UITheme | null;
  color_theme: ColorTheme | null;
  created_at: string;
  updated_at: string;
  last_login_at: string | null;
}

/**
 * Represents the data structure returned by the backend when checking authentication status or
 * when fetching the current user's details.
 */
export interface AuthenticatedUserResponse {
  user: User;
}

/**
 * Request payload for updating user theme.
 */
export interface UpdateUserThemePayload {
  ui_theme: UITheme;
  color_theme: ColorTheme;
}

/**
 * Request payload for updating user profile information.
 */
export interface UpdateUserProfilePayload {
  display_name?: string;
  avatar_url?: string;
  stats_privacy?: "private" | "friends_only" | "public";
}

/**
 * Represents a linked OAuth account for a user.
 */
export interface LinkedAccount {
  provider_name: string;
  provider_display_name?: string | null;
  provider_avatar_url?: string | null;
  proivder_email?: string | null;
}
