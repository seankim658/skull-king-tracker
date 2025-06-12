import { useState, useEffect } from "react";
import type { UserProfile } from "@/lib/api/types";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import { Link } from "react-router-dom";
import {
  getFullAvatarURL,
  getAvatarFallback,
  formatFriendshipStatus,
  errorExtract,
} from "@/lib/utils";
import { Badge } from "../ui/badge";
import {
  CalendarDays,
  Users,
  Edit3,
  UserPlus,
  UserCheck,
  UserX,
} from "lucide-react";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { toast } from "sonner";

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
}

export function ProfileHeader({ profile, isOwnProfile }: ProfileHeaderProps) {
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

  const handleSendFriendRequest = async () => {
    if (!profile || !profile.user_id) {
      toast.error("Cannot send request: User ID is missing");
      return;
    }

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

  const friendButton = () => {
    switch (friendshipStatus) {
      case "not_friends":
        return (
          <Button
            onClick={handleSendFriendRequest}
            disabled={isFriendActionLoading}
            className="cursor-pointer"
          >
            <UserPlus className="mr-2 h-4 w-4" />
            Add Friend
          </Button>
        );
      case "pending_sent_to_profile":
        return <Button disabled>Request Sent</Button>;
      case "pending_sent_to_viewer":
        return <Button disabled>Respond to Request</Button>;
      case "friends":
        return (
          <Button variant="secondary" disabled>
            <UserCheck className="mr-2 h-4 w-4" />
            Friends
          </Button>
        );
      case "blocked_by_viewer":
        return (
          <Button variant="destructive" disabled>
            <UserX className="mr-2 h-4 w-4" />
            Blocked
          </Button>
        );
      default:
        return null;
    }
  };

  return (
    <div className="bg-card p-6 rounded-lg shadow-md">
      <div className="flex flex-col sm:flex-row items-center space-y-4 sm:space-y-0 sm:space-x-6">
        <Avatar className="h-24 w-24 text-3xl">
          <AvatarImage src={avatarUrl} alt={displayName} />
          <AvatarFallback>{getAvatarFallback(displayName)}</AvatarFallback>
        </Avatar>
        <div className="flex-grow text-center sm:text-left">
          <h1 className="text-3xl font-bold">{displayName}</h1>
          <p className="text-muted-foreground">@{profile.username}</p>
          <div className="mt-2 flex flex-wrap justify-center sm:justify-start gap-2 text-sm text-muted-foreground">
            <span className="flex items-center">
              <CalendarDays className="mr-1.5 h-4 w-4" /> Joined on {joinDate}
            </span>
            <span className="flex items-center">
              <Users className="mr-1.5 h-4 w-4" /> {profile.friend_count}{" "}
              Friends
            </span>
          </div>
          {friendshipStatus !== "self" &&
            friendshipStatus !== "viewer_not_authenticated" && (
              <Badge variant="secondary" className="mt-3">
                {formatFriendshipStatus(friendshipStatus)}
              </Badge>
            )}
        </div>
        <div className="flex-shrink-0">
          {isOwnProfile ? (
            <Button asChild variant="outline" className="cursor-pointer">
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
  );
}
