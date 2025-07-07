import type { FriendshipStatus, UserProfile } from "@/lib/api/types";
import { UserAvatar } from "../ui/user-avatar";
import { Button } from "../ui/button";
import { Link } from "react-router-dom";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { formatFriendshipStatus } from "@/lib/utils";
import { MoreVertical, UserCheck, Users2 } from "lucide-react";
import { Badge } from "../ui/badge";
import { useConfirm } from "@/hooks/use-confirm";
import { CalendarDays, Users, Edit3, UserPlus, UserX } from "lucide-react";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { useSubmit } from "@/hooks/use-submit";
import { Tooltip, TooltipTrigger, TooltipContent } from "../ui/tooltip";

const friendButtonIconStyling = "mr-2 h-4 w-4";

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  onActionSuccess: () => void;
}

export function ProfileHeader({
  profile,
  isOwnProfile,
  onActionSuccess,
}: ProfileHeaderProps) {
  const confirm = useConfirm();

  const displayName = profile.display_name || profile.username;
  const joinDate = new Date(profile.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const { submit: sendFriendRequest, isLoading: isSendingRequest } = useSubmit(
    friendshipAPI.sendRequest,
    {
      actionVerb: "Sending friend request",
      successMessage: "Friend request sent",
      onSuccess: onActionSuccess,
    },
  );

  const { submit: unfriendUser, isLoading: isUnfriending } = useSubmit(
    friendshipAPI.unfriend,
    {
      actionVerb: "Submitting unfriend request",
      successMessage: `You are no longer friends with ${displayName}`,
      onSuccess: onActionSuccess,
    },
  );

  const { submit: cancelFriendRequest, isLoading: isCancelling } = useSubmit(
    friendshipAPI.cancelRequest,
    {
      actionVerb: "Canelling friend request",
      successMessage: "Friend request cancelled",
      onSuccess: onActionSuccess,
    },
  );

  const { submit: blockUser, isLoading: isBlocking } = useSubmit(
    friendshipAPI.blockUser,
    {
      actionVerb: "Blocking user",
      successMessage: `Blocked ${displayName}`,
      onSuccess: onActionSuccess,
    },
  );

  const { submit: unblockUser, isLoading: isUnblocking } = useSubmit(
    friendshipAPI.unblockUser,
    {
      actionVerb: "Unblocking user",
      successMessage: `Unblocked ${displayName}`,
      onSuccess: onActionSuccess,
    },
  );

  const friendshipStatus: FriendshipStatus =
    profile.friendship_status_with_viewer;

  const handleUnfriend = async () => {
    const ok = await confirm({
      title: "Unfriend user?",
      description: `Are you sure you want to unfriend ${displayName}`,
    });
    if (ok) unfriendUser(profile.user_id);
  };

  const handleBlock = async () => {
    const ok = await confirm({
      title: "Block user?",
      description: `Are you sure you want to block ${displayName}?`,
    });
    if (ok) blockUser(profile.user_id);
  };

  const handleUnblock = () => {
    unblockUser(profile.user_id);
  };

  const isLoading =
    isSendingRequest ||
    isUnfriending ||
    isCancelling ||
    isBlocking ||
    isUnblocking;

  const friendButton = () => {
    switch (friendshipStatus) {
      case "not_friends":
        return (
          <Button
            onClick={() => sendFriendRequest(profile.user_id)}
            disabled={isLoading}
            className="cursor-pointer"
          >
            <UserPlus className={friendButtonIconStyling} /> Add Friend
          </Button>
        );
      case "pending_sent_to_profile":
        return (
          <Button
            variant="outline"
            onClick={() => cancelFriendRequest(profile.user_id)}
            disabled={isLoading}
            className="cursor-pointer"
          >
            Cancel Request
          </Button>
        );
      case "pending_sent_to_viewer":
        return <Button disabled>Respond to Request</Button>;
      case "friends":
        return (
          <Button
            variant="destructive"
            onClick={handleUnfriend}
            disabled={isLoading}
            className="cursor-pointer"
          >
            <UserX className={friendButtonIconStyling} /> Unfriend
          </Button>
        );
      case "blocked_by_viewer":
        return (
          <Button
            variant="destructive"
            onClick={handleUnfriend}
            disabled={isLoading}
            className="cursor-pointer"
          >
            <UserCheck className={friendButtonIconStyling} /> Unblock
          </Button>
        );
      default:
        return null;
    }
  };

  return (
    <div className="relative bg-card p-6 rounded-lg shadow-md">
      {!isOwnProfile && (
        <div className="absolute top-2 right-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="cursor-pointer">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {friendshipStatus === "blocked_by_viewer" ? (
                <DropdownMenuItem onSelect={handleUnblock}>
                  Unblock User
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem
                  className="text-destructive cursor-pointer"
                  onSelect={handleBlock}
                >
                  Block User
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}

      <div className="flex flex-col gap-4">
        <div className="flex flex-row items-center gap-4">
          <UserAvatar
            displayName={displayName}
            avatarUrl={profile.avatar_url}
            className="h-20 w-20 text-2xl flex-shrink-0"
          />
          <div className="flex-grow text-left">
            <h1 className="text-2xl font-bold">{displayName}</h1>
            <p className="text-muted-foreground">@{profile.username}</p>
            <div className="mt-1 flex flex-col items-start text-sm text-muted-foreground">
              <span className="flex items-center">
                <Users className="mr-1.5 h-4 w-4" /> {profile.friend_count}{" "}
                Friends
              </span>
              {profile.mutual_friend_count &&
                profile.mutual_friend_count > 0 && (
                  <span className="flex items-center">
                    <Users2 className="mr-1.5 h-4 w-4" />
                    {
                      profile.mutual_friend_count
                    } mutual friend
                    {profile.mutual_friend_count > 1 ? "s" : ""}
                  </span>
                )}
              <span className="flex items-center">
                <CalendarDays className="mr-1.5 h-4 w-4" /> Joined on {joinDate}
              </span>
            </div>
          </div>
        </div>

        <div className="flex w-full items-center justify-between">
          <span>
            {friendshipStatus !== "self" &&
              friendshipStatus !== "viewer_not_authenticated" && (
                <Tooltip>
                  <TooltipTrigger>
                    <Badge variant="secondary" className="mt-3">
                      {formatFriendshipStatus(friendshipStatus)}
                    </Badge>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>
                      {friendshipStatus === "pending_sent_to_viewer"
                        ? "This user sent you a friend request. Check your notifications to respond."
                        : `Your friendship status with ${displayName}.`}
                    </p>
                  </TooltipContent>
                </Tooltip>
              )}
          </span>

          <div>
            {isOwnProfile ? (
              <Button
                asChild
                variant="outline"
                className="cursor-pointer w-full sm:w-auto"
              >
                <Link to="/settings">
                  <Edit3 className="mr-2 h-4 w-4" />
                  Edit Profile
                </Link>
              </Button>
            ) : (
              friendButton()
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
