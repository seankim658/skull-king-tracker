import React from "react";
import { useEffect, useCallback, useMemo } from "react";
import { Link } from "react-router-dom";
import { Bell, MailOpen, Mail } from "lucide-react";
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
import { Avatar, AvatarFallback, AvatarImage } from "./avatar";
import { notificationAPI } from "@/lib/api/service/notification";
import { friendshipAPI } from "@/lib/api/service/friendship";
import type { Notification } from "@/lib/api/types";
import {
  getFullAvatarURL,
  getAvatarFallback,
  errorExtract,
  cn,
} from "@/lib/utils";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { useApi } from "@/hooks/use-api";
import { useSubmit } from "@/hooks/use-submit";
import { SkeletonList } from "./skeleton-list";

export function NotificationBell() {
  const { isAuthenticated } = useAuth();

  const handleFetchError = useCallback((e: Error) => {
    console.error(`Failed to fetch notifications: ${errorExtract(e, "")}`);
  }, []);

  const apiOptions = useMemo(
    () => ({
      onError: handleFetchError,
    }),
    [handleFetchError],
  );

  const {
    data: notifications,
    isLoading,
    request: fetchNotifications,
  } = useApi(notificationAPI.getNotifications, apiOptions);

  useEffect(() => {
    if (isAuthenticated) {
      fetchNotifications();
      const interval = setInterval(fetchNotifications, 60000);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated, fetchNotifications]);

  const { submit: markAsRead, isLoading: isMarkingRead } = useSubmit(
    notificationAPI.markAsRead,
    {
      actionVerb: "Marking as read",
      onSuccess: () => fetchNotifications(),
    },
  );
  const { submit: markAsUnread, isLoading: isMarkingUnread } = useSubmit(
    notificationAPI.markAsUnread,
    {
      actionVerb: "Marking as unread",
      onSuccess: () => fetchNotifications(),
    },
  );

  const handleUpdateReadStatus = (notification: Notification) => {
    if (notification.is_read) {
      markAsUnread(notification.notification_id);
    } else {
      markAsRead(notification.notification_id);
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
      await notificationAPI.markAsRead(notification.notification_id);
      return { success: true };
    },
    [],
  );

  const { submit: handleResponse, isLoading: isResponding } = useSubmit(
    respondAndMarkRead,
    {
      actionVerb: "Responding to friend request",
      onSuccess: (_data, _notification, response) => {
        toast.success(`Friend request ${response}`);
        fetchNotifications();
      },
    },
  );

  if (!isAuthenticated) return null;

  const unreadCount = notifications?.filter((n) => !n.is_read).length || 0;

  const isActionLoading = isMarkingRead || isMarkingUnread || isResponding;

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
          <DropdownMenuLabel>Notifications</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {isLoading ? (
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
                    key={notif.notification_id}
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
                        <Avatar className="h-9 w-9">
                          <AvatarImage
                            src={getFullAvatarURL(notif.actor.avatar_url)}
                            alt={notif.actor.username}
                          />
                          <AvatarFallback>
                            {getAvatarFallback(
                              notif.actor.display_name || notif.actor.username,
                            )}
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex-1">
                          <p className="text-sm">{notif.message}</p>
                          <p className="text-xs text-muted-foreground">
                            {new Date(notif.created_at).toLocaleString()}
                          </p>
                        </div>
                      </Link>

                      <div className="flex-shrink-0">
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
