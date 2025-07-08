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
import { MoreVertical, UserRoundCheck } from "lucide-react";
import { useConfirm } from "@/hooks/use-confirm";
import { CalendarDays, Users, Edit3 } from "lucide-react";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { useSubmit } from "@/hooks/use-submit";
import { FriendshipActionButton } from "./friendship-button";

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  onActionSuccess: () => void;
  onOpenFriendsList: () => void;
}

export function ProfileHeader({
  profile,
  isOwnProfile,
  onActionSuccess,
  onOpenFriendsList,
}: ProfileHeaderProps) {
  const confirm = useConfirm();
  const displayName = profile.display_name || profile.username;
  const joinDate = new Date(profile.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const friendshipStatus: FriendshipStatus =
    profile.friendship_status_with_viewer;

  const { submit: blockUser, isLoading: isBlocking } = useSubmit(
    friendshipAPI.blockUser,
    {
      actionVerb: "Blocking user",
      successMessage: `Blocked ${displayName}`,
      onSuccess: onActionSuccess,
    },
  );

  const handleBlock = async () => {
    const ok = await confirm({
      title: "Block user?",
      description: `Are you sure you want to block ${displayName}?`,
    });
    if (ok) blockUser(profile.user_id);
  };

  return (
    <div className="relative bg-card p-6 rounded-lg shadow-md">
      {!isOwnProfile && friendshipStatus !== "blocked_by_viewer" && (
        <div className="absolute top-2 right-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="cursor-pointer"
                disabled={isBlocking}
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                className="text-destructive cursor-pointer"
                onSelect={handleBlock}
              >
                Block User
              </DropdownMenuItem>
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
              <span
                className="flex items-center hover:underline cursor-pointer"
                onClick={onOpenFriendsList}
                role="button"
              >
                {" "}
                <Users className="mr-1.5 h-4 w-4" /> {profile.friend_count}{" "}
                Friends
              </span>

              {profile.mutual_friend_count != null && (
                <span className="flex items-center">
                  <UserRoundCheck className="mr-1.5 h-4 w-4" />
                  {
                    profile.mutual_friend_count
                  } Mutual friend
                  {profile.mutual_friend_count !== 1 ? "s" : ""}
                </span>
              )}

              <span className="flex items-center">
                <CalendarDays className="mr-1.5 h-4 w-4" /> Joined on {joinDate}
              </span>
            </div>
          </div>
        </div>
        <div className="flex w-full items-center justify-between">
          <span></span>
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
              <FriendshipActionButton
                targetUser={{ userId: profile.user_id, displayName }}
                status={profile.friendship_status_with_viewer}
                onSuccess={onActionSuccess}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
