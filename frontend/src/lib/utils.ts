import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import type { FriendshipStatus } from "./api/types";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function errorExtract(error: any, defaultMsg: string): string {
  return error instanceof Error ? error.message : defaultMsg;
}

export function getFullAvatarURL(
  avatarPath: string | null | undefined,
): string {
  if (!avatarPath) {
    return "";
  }
  if (avatarPath.startsWith("http://") || avatarPath.startsWith("https://")) {
    return avatarPath;
  }
  const baseURL = import.meta.env.VITE_BACKEND_ASSET_BASE_URL || "";
  return `${baseURL}${avatarPath}`;
}

export function getAvatarFallback(displayName: string): string {
  return displayName.substring(0, 2).toUpperCase();
}

export function formatFriendshipStatus(status: FriendshipStatus): string {
  switch (status) {
    case "self":
      return "Your profile";
    case "friends":
      return "Friends";
    case "not_friends":
      return "Not friends";
    case "pending_sent_to_viewer":
      return "Friend request received";
    case "pending_sent_to_profile":
      return "Friend request sent";
    case "blocked_by_viewer":
      return "You blocked this user";
    case "blocked_by_profile_user":
      return "You are blocked by this user";
    case "viewer_not_authenticated":
      return "";
    default:
      return "Unknown";
  }
}
