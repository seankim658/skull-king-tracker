export interface UpdateReportStatusPayload {
  status: "payload" | "dismissed";
}

export interface SendNotificationPayload {
  message: string;
  user_ids: string[];
  is_broadcast: boolean;
}

export interface ReportUserPayload {
  reason: string;
}

export interface BanUserRequest {
  reason: string;
}
