import type { ComponentProps } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "./avatar";
import { getAvatarFallback, getFullAvatarURL, cn } from "@/lib/utils";

interface UserAvatarProps extends ComponentProps<typeof Avatar> {
  displayName: string;
  avatarUrl?: string | null;
  className?: string;
}

export function UserAvatar({
  displayName,
  avatarUrl,
  className,
  ...props
}: UserAvatarProps) {
  const fullUrl = getFullAvatarURL(avatarUrl);

  return (
    <Avatar className={cn("h-10 w-10", className)} {...props}>
      {fullUrl && <AvatarImage src={fullUrl} alt={displayName} />}
      <AvatarFallback>{getAvatarFallback(displayName)}</AvatarFallback>
    </Avatar>
  );
}
