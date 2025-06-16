import { useState, useEffect } from "react";
import type { UserProfile } from "@/lib/api/types";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import { Link } from "react-router-dom";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import {
  getFullAvatarURL,
  getAvatarFallback,
  formatFriendshipStatus,
  errorExtract,
} from "@/lib/utils";
import { MoreVertical, UserCheck } from "lucide-react";
import { Badge } from "../ui/badge";
import { useConfirm } from "@/hooks/use-confirm";
import { CalendarDays, Users, Edit3, UserPlus, UserX } from "lucide-react";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { toast } from "sonner";

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
  const [friendshipStatus, setFriendshipStatus] = useState(
    profile.friendship_status_with_viewer,
  );
  const [isFriendActionLoading, setIsFriendActionLoading] = useState(false);

  useEffect(() => {
    setFriendshipStatus(profile.friendship_status_with_viewer);
  }, [profile.friendship_status_with_viewer]);

  const avatarUrl = getFullAvatarURL(profile.avatar_url);
  const displayName = profile.display_name || profile.username;
  const joinDate = new Date(profile.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const handleFriendRequest = async () => {
    if (!profile || !profile.user_id) return;

    setIsFriendActionLoading(true);
    const toastId = toast.loading("Sending friend request...");
    try {
      const response = await friendshipAPI.sendRequest(profile.user_id);
      if (response.success) {
        toast.success("Friend request sent", { id: toastId });
        setFriendshipStatus("pending_sent_to_profile");
      } else {
        const errMsg = "Failed to send friend request";
        toast.error(errMsg);
        console.error(`${errMsg}: ${response.message}`);
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to send friend request");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsFriendActionLoading(false);
    }
  };

  const handleUnfriendRequest = async () => {
    if (!profile || !profile.user_id) return;

    const isConfirmed = await confirm({
      title: "Unfriend User?",
      description: `Are you sure you want to unfriend ${displayName}?`,
    });
    if (!isConfirmed) {
      return;
    }

    setIsFriendActionLoading(true);
    const toastId = toast.loading("Unfriending...");
    try {
      const response = await friendshipAPI.unfriend(profile.user_id);
      if (response.success) {
        toast.success(`You are no longer friends with ${displayName}`, {
          id: toastId,
        });
        setFriendshipStatus("not_friends");
      } else {
        const errMsg = response.message || "Failed to unfriend";
        toast.error(errMsg, { id: toastId });
        console.error(errMsg);
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to unfriend user");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsFriendActionLoading(false);
    }
  };

  const handleCancelRequest = async () => {
    if (!profile || !profile.user_id) return;

    setIsFriendActionLoading(true);
    const toastId = toast.loading("Canceling friend request...");
    try {
      const response = await friendshipAPI.cancelRequest(profile.user_id);
      if (response.success) {
        toast.success("Friend request canceled", { id: toastId });
        setFriendshipStatus("not_friends");
      } else {
        const errMsg = response.message || "Failed to cancel friend request";
        toast.error(errMsg);
        console.error(errMsg);
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to cancel friend request");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsFriendActionLoading(false);
    }
  };

  const handleBlockUser = async () => {
    if (!profile || !profile.user_id) return;

    const isConfirmed = await confirm({
      title: "Block User?",
      description: `Are you sure you want to block ${displayName}?`,
    });
    if (!isConfirmed) {
      return;
    }

    setIsFriendActionLoading(true);
    const toastId = toast.loading("Blocking user...");
    try {
      const response = await friendshipAPI.blockUser(profile.user_id);
      if (response.success) {
        toast.success(`Blocked ${displayName}`, { id: toastId });
        setFriendshipStatus("blocked_by_viewer");
      } else {
        const errMsg = response.message || "Failed to block user";
        toast.error(errMsg, { id: toastId });
        console.error(errMsg);
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to block user");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsFriendActionLoading(false);
    }
  };

  const handleUnblockUser = async () => {
    if (!profile || !profile.user_id) return;

    setIsFriendActionLoading(true);
    const toastId = toast.loading("Unblocking user...");
    try {
      const response = await friendshipAPI.unblockUser(profile.user_id);
      if (response.success) {
        toast.success(`Unblocked ${displayName}`, { id: toastId });
        setFriendshipStatus("not_friends");
      } else {
        const errMsg = response.message || "Failed to unblock user";
        toast.error(errMsg, { id: toastId });
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to unblock user");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsFriendActionLoading(false);
    }
  };

  const friendButton = () => {
    switch (friendshipStatus) {
      case "not_friends":
        return (
          <Button
            onClick={handleFriendRequest}
            disabled={isFriendActionLoading}
            className="cursor-pointer"
          >
            <UserPlus className="mr-2 h-4 w-4" />
            Add Friend
          </Button>
        );
      case "pending_sent_to_profile":
        return (
          <Button
            variant="outline"
            onClick={handleCancelRequest}
            disabled={isFriendActionLoading}
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
            onClick={handleUnfriendRequest}
            disabled={isFriendActionLoading}
            className="cursor-pointer"
          >
            <UserX className="mr-2 h-4 w-4" />
            Unfriend
          </Button>
        );
      case "blocked_by_viewer":
        return (
          <Button
            variant="outline"
            onClick={handleUnblockUser}
            disabled={isFriendActionLoading}
            className="cursor-pointer"
          >
            <UserCheck className="mr-2 h-4 w-4" />
            Unblock
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
                <DropdownMenuItem onSelect={handleUnblockUser}>
                  Unblock User
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem
                  className="text-destructive cursor-pointer"
                  onSelect={handleBlockUser}
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
          <Avatar className="h-20 w-20 text-2xl flex-shrink-0">
            <AvatarImage src={avatarUrl} alt={displayName} />
            <AvatarFallback>{getAvatarFallback(displayName)}</AvatarFallback>
          </Avatar>
          <div className="flex-grow text-left">
            <h1 className="text-2xl font-bold">{displayName}</h1>
            <p className="text-muted-foreground">@{profile.username}</p>
            <div className="mt-1 flex flex-col items-start text-sm text-muted-foreground">
              <span className="flex items-center">
                <Users className="mr-1.5 h-4 w-4" /> {profile.friend_count}{" "}
                Friends
              </span>
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
                <Badge variant="secondary" className="mt-3">
                  {formatFriendshipStatus(friendshipStatus)}
                </Badge>
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
