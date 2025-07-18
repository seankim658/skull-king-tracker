import { useConfirm } from "@/hooks/use-confirm";
import { useSubmit } from "@/hooks/use-submit";
import { friendshipAPI } from "@/lib/api/service/friendship";
import type { FriendshipStatus } from "@/lib/api/types";
import { Button } from "../ui/button";
import { UserCheck, UserPlus, UserX } from "lucide-react";
import { formatFriendshipStatus } from "@/lib/utils";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/hooks/use-auth";

const iconStyling = "mr-2 h-4 w-4";

interface FriendshipActionButtonProps {
  targetUser: {
    userId: string;
    displayName: string;
  };
  status: FriendshipStatus;
  onSuccess: () => void;
}

export function FriendshipActionButton({
  targetUser,
  status,
  onSuccess,
}: FriendshipActionButtonProps) {
  const confirm = useConfirm();
  const queryClient = useQueryClient();
  const { user: authenticatedUser } = useAuth();

  const handleActionSuccess = () => {
    queryClient.invalidateQueries({
      queryKey: ["userProfile", targetUser.userId],
    });
    if (authenticatedUser) {
      queryClient.invalidateQueries({
        queryKey: ["userProfile", authenticatedUser.user_id],
      });
    }
    queryClient.invalidateQueries({
      queryKey: ["friendsList", authenticatedUser?.user_id],
    });
    queryClient.invalidateQueries({
      queryKey: ["friendsList", targetUser.userId],
    });
    onSuccess();
  };

  const { submit: sendFriendRequest, isLoading: isSending } = useSubmit(
    friendshipAPI.sendRequest,
    {
      actionVerb: "Sending request",
      successMessage: "Friend request sent",
      onSuccess: handleActionSuccess,
    },
  );
  const { submit: unfriendUser, isLoading: isUnfriending } = useSubmit(
    friendshipAPI.unfriend,
    {
      actionVerb: "Unfriending",
      successMessage: `Unfriended ${targetUser.displayName}`,
      onSuccess: handleActionSuccess,
    },
  );
  const { submit: cancelFriendRequest, isLoading: isCancelling } = useSubmit(
    friendshipAPI.cancelRequest,
    {
      actionVerb: "Cancelling request",
      successMessage: "Friend request cancelled",
      onSuccess: handleActionSuccess,
    },
  );
  const { submit: unblockUser, isLoading: isUnblocking } = useSubmit(
    friendshipAPI.unblockUser,
    {
      actionVerb: "Unblocking user",
      successMessage: `Unblocked ${targetUser.displayName}`,
      onSuccess: handleActionSuccess,
    },
  );

  const isLoading = isSending || isUnfriending || isCancelling || isUnblocking;

  const handleUnfriend = async () => {
    const ok = await confirm({
      title: "Unfriend user?",
      description: `Are you sure you want to unfriend ${targetUser.displayName}?`,
    });
    if (ok) unfriendUser(targetUser.userId);
  };

  switch (status) {
    case "not_friends":
      return (
        <Button
          onClick={() => sendFriendRequest(targetUser.userId)}
          disabled={isLoading}
          className="cursor-pointer"
        >
          <UserPlus className={iconStyling} /> Add Friend
        </Button>
      );
    case "pending_sent_to_profile":
      return (
        <Button
          variant="outline"
          onClick={() => cancelFriendRequest(targetUser.userId)}
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
          <UserCheck className={iconStyling} /> Unfriend
        </Button>
      );
    case "blocked_by_viewer":
      return (
        <Button
          variant="destructive"
          onClick={() => unblockUser(targetUser.userId)}
          disabled={isLoading}
          className="cursor-pointer"
        >
          <UserX className={iconStyling} /> Unblock
        </Button>
      );
    default:
      return (
        <Button variant="secondary" className="cursor-pointer" disabled>
          {formatFriendshipStatus(status)}
        </Button>
      );
  }
}
