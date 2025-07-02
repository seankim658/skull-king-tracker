import { client } from "../client";
import type { ApiResponse, Notification } from "../types";

export const notificationAPI = {
  /**
   * Fetches the current user's notifications.
   */
  getNotifications: (): Promise<ApiResponse<Notification[]>> =>
    client<Notification[]>("/notifications", {
      method: "GET",
    }),

  /**
   * Marks a specific notification as read.
   * @param notificationId - The ID of the notification to mark as read
   */
  markAsRead: (notificationId: string): Promise<ApiResponse<null>> =>
    client<null>(`/notifications/${notificationId}/read`, {
      method: "PUT",
    }),

  /**
   * Marks a specific notification as unread.
   * @param notificationId - The ID of the notification to mark as unread
   */
  markAsUnread: (notificationId: string): Promise<ApiResponse<null>> =>
    client<null>(`/notifications/${notificationId}/read`, {
      method: "DELETE",
    }),
};
