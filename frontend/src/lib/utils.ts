import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

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
