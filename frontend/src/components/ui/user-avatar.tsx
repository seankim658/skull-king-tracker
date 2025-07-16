import type { ComponentProps } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "./avatar";
import { getAvatarFallback, getFullAvatarURL, cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";

interface UserAvatarProps extends ComponentProps<typeof Avatar> {
  userId?: string | null;
  displayName: string;
  avatarUrl?: string | null;
  updatedAt?: string | null;
  className?: string;
}

export function UserAvatar({
  userId,
  displayName,
  avatarUrl,
  updatedAt,
  className,
  ...props
}: UserAvatarProps) {
  const { user: authUser } = useAuth();
  let finalTimestamp = updatedAt;

  // If this avatar is for the logged-in user, use the fresh timestamp from our auth context.
  // This ensures the user's own avatar is always up-to-date everywhere.
  if (authUser && userId === authUser.user_id) {
    finalTimestamp = authUser.updated_at;
  }

  const fullUrl = getFullAvatarURL(avatarUrl, finalTimestamp);

  return (
    <Avatar className={cn("h-10 w-10", className)} {...props}>
      {fullUrl && <AvatarImage src={fullUrl} alt={displayName} />}
      <AvatarFallback>{getAvatarFallback(displayName)}</AvatarFallback>
    </Avatar>
  );
}
