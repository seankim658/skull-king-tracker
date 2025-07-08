import { useNavigate } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { userAPI } from "@/lib/api/service/user";
import type { UserProfile } from "@/lib/api/types";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "../ui/dialog";
import { SkeletonList } from "../ui/skeleton-list";
import { UserAvatar } from "../ui/user-avatar";
import { FriendshipActionButton } from "./friendship-button";

interface FriendsListModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: UserProfile;
}

export function FriendsListModal({
  isOpen,
  onClose,
  profile,
}: FriendsListModalProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: friends, isLoading } = useQuery({
    queryKey: ["friendsList", profile.user_id],
    queryFn: () => userAPI.getFriendsList(profile.user_id),
    enabled: isOpen,
  });

  const handleActionSuccess = () => {
    queryClient.invalidateQueries({
      queryKey: ["friendsList", profile.user_id],
    });
  };

  const handleUserClick = (userId: string) => {
    onClose();
    navigate(`/users/${userId}`);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Friends</DialogTitle>
        </DialogHeader>
        <div className="max-h-[60vh] overflow-y-auto -mx-6 px-6">
          {isLoading && <SkeletonList count={5} />}
          {!isLoading &&
            friends?.data?.map((friend) => (
              <div
                key={friend.user_id}
                className="flex items-center space-x-4 py-2"
              >
                <button
                  className="flex items-center space-x-4 flex-1"
                  onClick={() => handleUserClick(friend.user_id)}
                >
                  <UserAvatar
                    displayName={friend.display_name || friend.username}
                    avatarUrl={friend.avatar_url}
                  />
                  <div className="text-left">
                    <p className="text-sm font-medium leading-none hover:underline cursor-pointer">
                      {friend.display_name || friend.username}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      @{friend.username}
                    </p>
                  </div>
                </button>
                <FriendshipActionButton
                  targetUser={{
                    userId: friend.user_id,
                    displayName: friend.display_name || friend.username,
                  }}
                  status={friend.friendship_status_with_viewer}
                  onSuccess={handleActionSuccess}
                />
              </div>
            ))}
          {!isLoading && friends?.data?.length === 0 && (
            <p className="text-center text-sm text-muted-foreground py-8">
              No friends to display.
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
