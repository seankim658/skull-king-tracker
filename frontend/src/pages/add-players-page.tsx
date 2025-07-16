import { useState, useRef, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { gameAPI } from "@/lib/api/service/game";
import type { GamePlayerResponse, UserSearchItem } from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UserAvatar } from "@/components/ui/user-avatar";
import { PlusCircle, UserPlus, X, Rocket, UserMinus } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { useSubmit } from "@/hooks/use-submit";
import { useConfirm } from "@/hooks/use-confirm";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useQuery,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { InfoTooltip } from "@/components/ui/info-tooltip";
import { useDebounce } from "@/hooks/use-debounce";

export function AddPlayersPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const queryClient = useQueryClient();

  const [guestName, setGuestName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  const {
    data: friendsData,
    fetchNextPage,
    hasNextPage,
    isLoading: isLoadingFriends,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["friends", debouncedSearchQuery],
    queryFn: ({ pageParam = 1 }) =>
      friendshipAPI.getFriends({
        query: debouncedSearchQuery,
        page: pageParam,
        pageSize: 15,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const pagination = lastPage.data?.pagination;
      if (!pagination || pagination.current_page >= pagination.total_pages) {
        return undefined;
      }
      return pagination.current_page + 1;
    },
    enabled:
      debouncedSearchQuery.trim().length === 0 ||
      debouncedSearchQuery.trim().length > 2,
  });

  const friends =
    friendsData?.pages.flatMap((page) => page.data?.users ?? []) ?? [];

  const { data: players, isLoading: isLoadingPlayers } = useQuery({
    queryKey: ["gamePlayers", gameId],
    queryFn: async () => {
      const response = await gameAPI.getGamePlayers(gameId!);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Could not fetch game players");
      }
      return response.data;
    },
    enabled: !!gameId,
  });

  const observer = useRef<IntersectionObserver | null>(null);
  const lastFriendElementRef = useCallback(
    (node: HTMLDivElement) => {
      if (isLoadingFriends || isFetchingNextPage) return;
      if (observer.current) observer.current.disconnect();
      observer.current = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting && hasNextPage) {
          fetchNextPage();
        }
      });
      if (node) observer.current.observe(node);
    },
    [isLoadingFriends, hasNextPage, isFetchingNextPage, fetchNextPage],
  );

  const { submit: addPlayerSubmit, isLoading: isAddingPlayer } = useSubmit(
    gameAPI.addPlayerToGame,
    {
      actionVerb: "Adding player",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["gamePlayers", gameId] });
        queryClient.invalidateQueries({ queryKey: ["pendingGames"] });
      },
    },
  );

  const { submit: removePlayerSubmit, isLoading: isRemovingPlayer } = useSubmit(
    gameAPI.removePlayerFromGame,
    {
      actionVerb: "Removing player",
      successMessage: "Player removed",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["gamePlayers", gameId] });
        queryClient.invalidateQueries({ queryKey: ["pendingGames"] });
      },
    },
  );

  const handleAddFriend = (userId: string) => {
    if (!gameId) return;
    addPlayerSubmit(gameId, {
      user_id: userId,
      seating_order: (players?.length || 0) + 1,
    });
  };

  const handleAddGuest = (e: React.FormEvent) => {
    e.preventDefault();
    if (!gameId || !guestName.trim()) return;
    addPlayerSubmit(gameId, {
      guest_name: guestName.trim(),
      seating_order: (players?.length || 0) + 1,
    });
    setGuestName("");
  };

  const handleRemovePlayer = async (player: GamePlayerResponse) => {
    if (!gameId) return;
    const isConfirmed = await confirm({
      title: "Remove player?",
      description: `Are you sure you want to remove ${player.display_name} from the game?`,
      confirmText: "Remove",
    });
    if (isConfirmed) {
      removePlayerSubmit(gameId, player.game_player_id);
    }
  };

  const handleRemoveFriend = (userIdToRemove: string) => {
    const playerInGame = players?.find((p) => p.user_id === userIdToRemove);
    if (playerInGame) {
      handleRemovePlayer(playerInGame);
    }
  };

  const handleProceed = () => {
    if ((players?.length ?? 0) < 2) {
      toast.error("A game requires at least 2 players to proceed");
      return;
    }
    navigate(`/game/${gameId}/setup`);
  };

  const alreadyAdded = (userId: string) =>
    players?.some((p) => p.user_id === userId);
  const isActionLoading = isAddingPlayer || isRemovingPlayer;
  const playerList = players || [];

  return (
    <div className="container mx-auto max-w-4xl p-4 md:p-6 space-y-8">
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">Game Lobby</CardTitle>
          <CardDescription>Add players to your game</CardDescription>
        </CardHeader>
        <CardContent className="grid md:grid-cols-2 gap-8">
          {/* Left side: Current Players */}
          <div>
            <h3 className="text-lg font-semibold mb-4">
              Current Players ({playerList.length})
            </h3>
            <div className="space-y-3">
              {isLoadingPlayers ? (
                <div className="space-y-3">
                  {[...Array(2)].map((_, i) => (
                    <Skeleton key={i} className="h-14 w-full" />
                  ))}
                </div>
              ) : playerList.length > 0 ? (
                playerList.map((player) => (
                  <div
                    key={player.game_player_id}
                    className="flex items-center justify-between bg-muted p-2 rounded-md"
                  >
                    <div className="flex items-center gap-3">
                      <UserAvatar
                        userId={player.user_id}
                        displayName={player.display_name}
                        avatarUrl={player.avatar_url}
                        className="h-9 w-9"
                      />
                      <span className="font-medium">{player.display_name}</span>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 cursor-pointer"
                      onClick={() => handleRemovePlayer(player)}
                      disabled={isActionLoading}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ))
              ) : (
                <p className="text-sm text-muted-foreground text-center py-4">
                  No players added yet
                </p>
              )}
            </div>
          </div>

          {/* Right Side: Add Players */}
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-2">Add a Friend</h3>
              <div className="relative mb-4">
                <Input
                  placeholder="Search friends..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="mb-4"
                />
                {searchQuery.length > 0 && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute right-2 top-1/2 -translate-y-1/2 h-7 w-7 rounded-full cursor-pointer"
                    onClick={() => setSearchQuery("")}
                  >
                    <X className="h-4 w-4" />
                    <span className="sr-only">Clear search</span>
                  </Button>
                )}
              </div>
              <div className="space-y-2 max-h-60 overflow-y-auto pr-2">
                {isLoadingFriends && !isFetchingNextPage ? (
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <Skeleton key={i} className="h-12 w-full" />
                    ))}
                  </div>
                ) : (
                  friends?.map((friend: UserSearchItem, index: number) => (
                    <div
                      key={friend.user_id}
                      ref={
                        index === friends.length - 1
                          ? lastFriendElementRef
                          : null
                      }
                      className="flex items-center gap-3"
                    >
                      <UserAvatar
                        userId={friend.user_id}
                        displayName={friend.display_name || friend.username}
                        avatarUrl={friend.avatar_url}
                        className="h-9 w-9"
                      />
                      <div className="flex-grow">
                        <p className="text-sm font-medium">
                          {friend.display_name || friend.username}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          @{friend.username}
                        </p>
                      </div>
                      {alreadyAdded(friend.user_id) ? (
                        <Button
                          size="sm"
                          variant="destructive"
                          className="cursor-pointer"
                          onClick={() => handleRemoveFriend(friend.user_id)}
                          disabled={isActionLoading}
                        >
                          <UserMinus className="h-4 w-4 mr-2" /> Remove
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          className="cursor-pointer"
                          variant="outline"
                          onClick={() => handleAddFriend(friend.user_id)}
                          disabled={
                            alreadyAdded(friend.user_id) || isActionLoading
                          }
                        >
                          <UserPlus className="h-4 w-4 mr-2" /> Add
                        </Button>
                      )}
                    </div>
                  ))
                )}
                {isFetchingNextPage && (
                  <p className="text-center text-sm text-muted-foreground py-2">
                    Loading more...
                  </p>
                )}
                {!isLoadingFriends && friends?.length === 0 && (
                  <p className="text-sm text-center text-muted-foreground py-2">
                    {" "}
                    No friends found.
                  </p>
                )}
              </div>
            </div>

            <Separator />

            <div>
              <div className="relative inline-flex items-center gap-2 mb-1">
                <h3 className="text-lg font-semibold mr-2">Add a Guest</h3>
                <InfoTooltip content="Guests are for players who do not have an account. Their scores will be tracked for this game, but their long-term stats will not be saved and they cannot be made scorekeeper." />
              </div>
              <form
                onSubmit={handleAddGuest}
                className="flex items-center gap-2"
              >
                <Label htmlFor="guest-name" className="sr-only">
                  Guest Name
                </Label>
                <Input
                  id="guest-name"
                  placeholder="Enter guest's name"
                  value={guestName}
                  onChange={(e) => setGuestName(e.target.value)}
                  disabled={isActionLoading}
                />
                <Button
                  type="submit"
                  className="cursor-pointer"
                  disabled={isActionLoading || !guestName.trim()}
                >
                  <PlusCircle className="h-4 w-4 mr-2" /> Add
                </Button>
              </form>
            </div>
          </div>
        </CardContent>
        <CardFooter className="pt-6">
          <Button
            size="lg"
            className="w-full md:w-auto cursor-pointer"
            disabled={playerList.length < 2 || isActionLoading}
            onClick={handleProceed}
            title={
              playerList.length > 2
                ? "A game requires at least 2 players"
                : "Proceed to game setup"
            }
          >
            <Rocket className="h-4 w-4 mr-2" />
            Proceed to Game Setup
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
