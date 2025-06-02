import type { UserProfile } from "@/lib/api/types";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Button } from "../ui/button";
import { Link } from "react-router-dom";
import {
  getFullAvatarURL,
  getAvatarFallback,
  formatFriendshipStatus,
} from "@/lib/utils";
import { Badge } from "../ui/badge";
import { CalendarDays, Users, Edit3 } from "lucide-react";

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
}

export function ProfileHeader({ profile, isOwnProfile }: ProfileHeaderProps) {
  const avatarUrl = getFullAvatarURL(profile.avatar_url);
  const displayName = profile.display_name || profile.username;
  const joinDate = new Date(profile.created_at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const friendShipStatusText = formatFriendshipStatus(
    profile.friendship_status_with_viewer,
  );

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
          {friendShipStatusText && (
            <Badge variant="secondary" className="mt-2">
              {friendShipStatusText}
            </Badge>
          )}
        </div>
        {isOwnProfile && (
          <Button asChild variant="outline">
            <Link to="/settings">
              <Edit3 className="mr-2 h-4 w-4" />
              Edit Profile
            </Link>
          </Button>
        )}
        {/* TODO : Friend button */}
      </div>
    </div>
  );
}
