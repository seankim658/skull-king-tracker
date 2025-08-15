import type { AdminUserView } from "@/lib/api/types";
import { useSubmit } from "@/hooks/use-submit";
import { adminAPI } from "@/lib/api/service/admin";
import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuSeparator,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { MoreHorizontal, Ban, ShieldCheck, User } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";

export function UserActions({
  user,
  onBan,
}: {
  user: AdminUserView;
  onBan: (user: AdminUserView) => void;
}) {
  const queryClient = useQueryClient();

  const { submit: unbanUser, isLoading } = useSubmit(adminAPI.unbanUser, {
    actionVerb: "Unbanning user",
    successMessage: `User ${
      user.display_name || user.username
    } has been unbanned.`,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["adminUsers"] });
      queryClient.invalidateQueries({
        queryKey: ["userProfile", user.user_id],
      });
    },
  });

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem asChild className="cursor-pointer">
          <Link to={`/users/${user.user_id}`}>
            <User className="mr-2 h-4 w-4" />
            View Profile
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {user.is_banned ? (
          <DropdownMenuItem
            onClick={() => unbanUser(user.user_id)}
            disabled={isLoading}
            className="cursor-pointer"
          >
            <ShieldCheck className="mr-2 h-4 w-4" />
            Unban User
          </DropdownMenuItem>
        ) : (
          <DropdownMenuItem
            onClick={() => onBan(user)}
            className="text-destructive cursor-pointer"
          >
            <Ban className="mr-2 h-4 w-4" />
            Ban User
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
