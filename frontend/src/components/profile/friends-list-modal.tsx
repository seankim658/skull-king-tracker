import { useNavigate } from "react-router-dom";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { userAPI } from "@/lib/api/service/user";
import type {
  ApiResponse,
  PaginatedFriendsListResponse,
  UserProfile,
} from "@/lib/api/types";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "../ui/dialog";
import { SkeletonList } from "../ui/skeleton-list";
import { UserAvatar } from "../ui/user-avatar";
import { FriendshipActionButton } from "./friendship-button";
import { useEffect, useRef } from "react";

interface FriendsListModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: UserProfile;
}

const PAGE_SIZE = 20;

export function FriendsListModal({
  isOpen,
  onClose,
  profile,
}: FriendsListModalProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const observerTarget = useRef(null);

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery<ApiResponse<PaginatedFriendsListResponse>>({
      queryKey: ["friendsList", profile.user_id],
      queryFn: ({ pageParam }) =>
        userAPI.getFriendsList(profile.user_id, pageParam as number, PAGE_SIZE),
      initialPageParam: 1,
      getNextPageParam: (lastPage) => {
        const { current_page, total_pages } = lastPage.data!.pagination;
        return current_page < total_pages ? current_page + 1 : undefined;
      },
      enabled: isOpen,
    });

  const friends = data?.pages.flatMap((page) => page.data?.friends ?? []) ?? [];

  const handleActionSuccess = () => {
    queryClient.invalidateQueries({
      queryKey: ["friendsList", profile.user_id],
    });
    friends.forEach((friend) => {
      queryClient.invalidateQueries({
        queryKey: ["userProfile", friend.user_id],
      });
    });
  };

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage();
        }
      },
      { threshold: 1 },
    );

    const target = observerTarget.current;
    if (target) {
      observer.observe(target);
    }

    return () => {
      if (target) {
        observer.unobserve(target);
      }
    };
  }, [observerTarget, hasNextPage, isFetchingNextPage, fetchNextPage]);

  const handleUserClick = (userId: string) => {
    onClose();
    navigate(`/users/${userId}`);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Friends of {profile.display_name}</DialogTitle>
        </DialogHeader>
        <div className="max-h-[60vh] overflow-y-auto -mx-6 px-6">
          {isLoading && <SkeletonList count={5} />}
          {!isLoading &&
            friends.map((friend) => (
              <div
                key={friend.user_id}
                className="flex items-center space-x-4 py-2"
              >
                <button
                  className="flex items-center space-x-4 flex-1"
                  onClick={() => handleUserClick(friend.user_id)}
                >
                  <UserAvatar
                    userId={friend.user_id}
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
          {!isLoading && friends.length === 0 && (
            <p className="text-center text-sm text-muted-foreground py-8">
              No friends to display.
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
