import React from "react";
import { useEffect, useCallback } from "react";
import { Link } from "react-router-dom";
import { Bell, MailOpen, Mail, Trash2, Sparkles } from "lucide-react";
import { Button } from "./button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "./tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import { Badge } from "./badge";
import { UserAvatar } from "./user-avatar";
import { notificationAPI } from "@/lib/api/service/notification";
import { friendshipAPI } from "@/lib/api/service/friendship";
import type {
  Notification,
  SSEEvent,
  SSEDeletedNotificationPayload,
} from "@/lib/api/types";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { useSubmit } from "@/hooks/use-submit";
import { SkeletonList } from "./skeleton-list";
import { useQueryClient } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { useConfirm } from "@/hooks/use-confirm";

export function NotificationBell() {
  const { isAuthenticated, user } = useAuth();
  const queryClient = useQueryClient();
  const confirm = useConfirm();

  const { data: notifications, isLoading: isLoadingInitial } = useQuery({
    queryKey: ["notifications"],
    queryFn: async () => {
      const response = await notificationAPI.getNotifications();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch notifications");
      }
      return response.data;
    },
    enabled: !!isAuthenticated,
    staleTime: 1000 * 60,
  });

  useEffect(() => {
    if (!isAuthenticated) {
      return;
    }

    const eventsUrl = `${import.meta.env.VITE_SSE_BASE_URL}/notifications/events`;
    console.log("Attempting to connect to SSE:", eventsUrl);

    const eventSource = new EventSource(eventsUrl, { withCredentials: true });

    eventSource.onopen = (event) => {
      console.log("SSE connection opened:", event);
    };

    eventSource.onmessage = (event) => {
      console.log("SSE message received:", event.data);
      try {
        const sseEvent: SSEEvent = JSON.parse(event.data);

        switch (sseEvent.event) {
          case "notification_created": {
            const newNotification = sseEvent.payload as Notification;
            toast.info(
              <span>
                New notification from{" "}
                <b>
                  {newNotification.actor.display_name ||
                    newNotification.actor.username}
                </b>
              </span>,
            );

            queryClient.setQueryData(
              ["notifications"],
              (oldData: Notification[] | undefined) => {
                if (!oldData) return [newNotification];
                if (
                  oldData.some(
                    (n) =>
                      n.notification_id === newNotification.notification_id,
                  )
                )
                  return oldData;
                return [newNotification, ...oldData];
              },
            );
            break;
          }

          case "notification_deleted": {
            const { notification_id } =
              sseEvent.payload as SSEDeletedNotificationPayload;
            queryClient.setQueryData(
              ["notifications"],
              (oldData: Notification[] | undefined) => {
                if (!oldData) return [];
                return oldData.filter(
                  (n) => n.notification_id !== notification_id,
                );
              },
            );
            break;
          }

          default:
            console.warn(`Unknown SSE event type: ${sseEvent.event}`);
            break;
        }
      } catch (e) {
        console.error("Failed to parse SSE notification data:", e);
      }
    };

    eventSource.onerror = (event) => {
      console.error("SSE connection error", event);
      console.log(`EventSource readystate:`, eventSource.readyState);
      console.log(`EventSource url:`, eventSource.url);

      if (eventSource.readyState === EventSource.CONNECTING) {
        console.log(`SSE is reconnecting...`);
      } else if (eventSource.readyState === EventSource.CLOSED) {
        console.log(`SSE connection closed`);
      }
    };

    eventSource.addEventListener("connected", (event) => {
      console.log("SSE connected event received:", event.data);
    });

    return () => {
      console.log("Closing SSE connection");
      eventSource.close();
    };
  }, [isAuthenticated, queryClient]);

  const { submit: markAsRead, isLoading: isMarkingRead } = useSubmit(
    notificationAPI.markAsRead,
    {
      actionVerb: "Marking as read",
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
    },
  );
  const { submit: markAsUnread, isLoading: isMarkingUnread } = useSubmit(
    notificationAPI.markAsUnread,
    {
      actionVerb: "Marking as unread",
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
    },
  );

  const { submit: deleteNotification, isLoading: isDeleting } = useSubmit(
    notificationAPI.deleteNotification,
    {
      actionVerb: "Deleting",
      successMessage: "Notification cleared.",
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
    },
  );

  const { submit: deleteAll, isLoading: isDeletingAll } = useSubmit(
    notificationAPI.deleteAllNotifications,
    {
      actionVerb: "Clearing all",
      successMessage: "All notifications cleared.",
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
    },
  );

  const handleUpdateReadStatus = (notification: Notification) => {
    if (notification.is_read) {
      markAsUnread(notification.notification_id);
    } else {
      markAsRead(notification.notification_id);
    }
  };

  const handleClearAll = async () => {
    const isConfirmed = await confirm({
      title: "Clear all notifications?",
      description: "This action cannot be undone.",
      confirmText: "Clear All",
    });
    if (isConfirmed) {
      deleteAll();
    }
  };

  const respondAndMarkRead = useCallback(
    async (notification: Notification, response: "accept" | "decline") => {
      if (!notification.friendship_id) {
        throw new Error("Cannot respond: friendship ID is missing");
      }
      await friendshipAPI.respondToRequest(
        notification.friendship_id,
        response,
      );
      return { success: true };
    },
    [],
  );

  const { submit: handleResponse, isLoading: isResponding } = useSubmit(
    respondAndMarkRead,
    {
      actionVerb: "Responding to friend request",
      onSuccess: (_data, notification, response) => {
        toast.success(`Friend request ${response}`);
        queryClient.invalidateQueries({ queryKey: ["notifications"] });
        queryClient.invalidateQueries({
          queryKey: ["userProfile", notification.actor.user_id],
        });
        if (user) {
          queryClient.invalidateQueries({
            queryKey: ["userProfile", user.user_id],
          });
        }
        queryClient.invalidateQueries({ queryKey: ["friendsList"] });
      },
    },
  );

  if (!isAuthenticated) return null;

  const unreadCount = notifications?.filter((n) => !n.is_read).length || 0;
  const isActionLoading =
    isMarkingRead ||
    isMarkingUnread ||
    isResponding ||
    isDeleting ||
    isDeletingAll;

  return (
    <TooltipProvider delayDuration={100}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="relative cursor-pointer"
          >
            <Bell className="h-5 w-5" />
            {unreadCount > 0 && (
              <Badge
                variant="destructive"
                className="absolute -top-1 -right-1 h-4 w-4 justify-center rounded-full p00 text-xs"
              >
                {unreadCount}
              </Badge>
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-96">
          <div className="flex items-center justify-between pr-2">
            <DropdownMenuLabel>Notifications</DropdownMenuLabel>
            {notifications && notifications.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleClearAll}
                disabled={isActionLoading}
                className="cursor-pointer"
              >
                <Sparkles className="mr-2 h-3 w-3" />
                Clear All
              </Button>
            )}
          </div>
          <DropdownMenuSeparator />
          {isLoadingInitial && !notifications ? (
            <SkeletonList count={2} />
          ) : notifications?.length === 0 ? (
            <p className="p-4 text-sm text-center text-muted-foreground">
              You have no new notifications.
            </p>
          ) : (
            <div className="max-h-[60vh] overflow-y-auto">
              {notifications?.map((notif, index) => (
                <React.Fragment key={notif.notification_id}>
                  <DropdownMenuItem
                    className={cn(
                      "flex flex-col items-start gap-2 p-3 cursor-default data-[highlighted]:bg-accent/50",
                      !notif.is_read && "bg-accent/40",
                    )}
                    onSelect={(e) => e.preventDefault()}
                  >
                    <div className="flex flex-row items-start justify-between w-full gap-3">
                      <div className="flex-shrink-0 pt-2">
                        {!notif.is_read && (
                          <span className="block h-2 w-2 rounded-full bg-primary" />
                        )}
                      </div>

                      <Link
                        to={`/users/${notif.actor.user_id}`}
                        className="flex-grow flex items-center gap-3"
                      >
                        <UserAvatar
                          displayName={
                            notif.actor.display_name || notif.actor.username
                          }
                          avatarUrl={notif.actor.avatar_url}
                          className="h-9 w-9"
                        />
                        <div className="flex-1">
                          <p className="text-sm">{notif.message}</p>
                          <p className="text-xs text-muted-foreground">
                            {new Date(notif.created_at).toLocaleString()}
                          </p>
                        </div>
                      </Link>

                      <div className="flex-shrink-0 flex items-center">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7 cursor-pointer"
                              onClick={() => handleUpdateReadStatus(notif)}
                              disabled={isActionLoading}
                            >
                              {notif.is_read ? (
                                <Mail className="h-4 w-4" />
                              ) : (
                                <MailOpen className="h-4 w-4" />
                              )}
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>
                              {notif.is_read
                                ? "Mark as unread"
                                : "Mark as read"}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7 text-destructive/80 hover:text-destructive cursor-pointer"
                              onClick={() =>
                                deleteNotification(notif.notification_id)
                              }
                              disabled={isActionLoading}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>Clear</p>
                          </TooltipContent>
                        </Tooltip>
                      </div>
                    </div>
                    {notif.type === "friend_request" && (
                      <div className="flex w-full justify-end gap-2 mt-1">
                        <Button
                          size="sm"
                          className="cursor-pointer"
                          onClick={() => handleResponse(notif, "accept")}
                          disabled={isActionLoading}
                        >
                          Accept
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          className="cursor-pointer"
                          onClick={() => handleResponse(notif, "decline")}
                          disabled={isActionLoading}
                        >
                          Decline
                        </Button>
                      </div>
                    )}
                  </DropdownMenuItem>
                  {index < notifications.length - 1 && (
                    <DropdownMenuSeparator />
                  )}
                </React.Fragment>
              ))}
            </div>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </TooltipProvider>
  );
}
