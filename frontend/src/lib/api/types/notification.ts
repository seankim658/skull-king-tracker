export interface NotificationActor {
  user_id: string;
  username: string;
  display_name?: string | null;
  avatar_url?: string | null;
}

export interface Notification {
  notification_id: string;
  type: string;
  actor: NotificationActor;
  message: string;
  is_read: boolean;
  friendship_id?: string | null;
  created_at: string;
}
