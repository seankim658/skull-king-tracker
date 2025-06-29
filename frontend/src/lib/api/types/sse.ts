export interface SSEEvent {
  event: "notification_created" | "notification_deleted";
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  payload: any;
}

export interface SSEDeletedNotificationPayload {
  notification_id: string;
}
