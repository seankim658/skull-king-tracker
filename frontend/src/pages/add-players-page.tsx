import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { friendshipAPI } from "@/lib/api/service/friendship";
import { gameAPI } from "@/lib/api/service/game";
import type { UserSearchItem, GamePlayerResponse } from "@/lib/api/types";
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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getFullAvatarURL, getAvatarFallback, errorExtract } from "@/lib/utils";
import { PlusCircle, UserPlus, X } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { useApi } from "@/hooks/use-api";
import { useSubmit } from "@/hooks/use-submit";

export function AddPlayersPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();

  const [players, setPlayers] = useState<GamePlayerResponse[]>([]);
  const [guestName, setGuestName] = useState("");

  const {
    data: friends,
    isLoading: isLoadingFriends,
    request: fetchFriends,
  } = useApi(friendshipAPI.getFriends);

  useEffect(() => {
    if (!gameId) {
      toast.error("Game ID is missing");
      navigate("/");
      return;
    }
    fetchFriends();
  }, [gameId, navigate, fetchFriends]);

  const { submit: addPlayer, isLoading: isAddingPlayer } = useSubmit(
    gameAPI.addPlayerToGame,
    {
      actionVerb: "Adding player",
      onSuccess: (newPlayer) => {
        if (newPlayer) setPlayers((prev) => [...prev, newPlayer]);
      },
    },
  );

  const { submit: addGuest, isLoading: isAddingGuest } = useSubmit(
    gameAPI.addPlayerToGame,
    {
      actionVerb: "Adding guest",
      onSuccess: (newPlayer) => {
        if (newPlayer) {
          setPlayers((prev) => [...prev, newPlayer]);
          setGuestName("");
        }
      },
    },
  );

  const handleAddPlayer = async (userId: string) => {
    if (!gameId) return;
    addPlayer(gameId, { user_id: userId, seating_order: players.length + 1 });
  };

  const handleAddGuest = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!gameId || !guestName.trim()) return;
    addGuest(gameId, {
      guest_name: guestName.trim(),
      seating_order: players.length + 1,
    });
  };

  // TODO Implement remove player logic
  const handleRemovePlayer = (gamePlayerId: string) => {
    toast.info("Removing player");
  };

  const alreadyAdded = (userId: string) =>
    players.some((p) => p.user_id === userId);

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
              Current Players ({players.length})
            </h3>
            <div className="space-y-3">
              {players.length > 0 ? (
                players.map((player) => (
                  <div
                    key={player.game_player_id}
                    className="flex items-center justify-between bg-muted p-2 rounded-md"
                  >
                    <div className="flex items-center gap-3"></div>
                    <Avatar className="h-9 w-9">
                      <AvatarImage src="" alt={player.display_naem} />
                      <AvatarFallback>
                        {getAvatarFallback(player.display_naem)}
                      </AvatarFallback>
                    </Avatar>
                    <span className="font-medium">{player.display_naem}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => handleRemovePlayer(player.game_player_id)}
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
              <div className="space-y-2 max-h-60 overflow-y-auto pr-2">
                {isLoadingFriends ? (
                  <p>Loading friends...</p>
                ) : (
                  friends?.map((friend) => (
                    <div
                      key={friend.user_id}
                      className="flex items-center gap-3"
                    >
                      <Avatar className="h-9 w-9">
                        <AvatarImage
                          src={getFullAvatarURL(friend.avatar_url)}
                          alt={friend.username}
                        />
                        <AvatarFallback>
                          {getAvatarFallback(
                            friend.display_name || friend.username,
                          )}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex-grow">
                        <p className="text-sm font-medium">
                          {friend.display_name || friend.username}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          @{friend.username}
                        </p>
                      </div>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleAddPlayer(friend.user_id)}
                        disabled={alreadyAdded(friend.user_id)}
                      >
                        <UserPlus className="h-4 w-4 mr-2" /> Add
                      </Button>
                    </div>
                  ))
                )}
              </div>
            </div>

            <Separator />

            <div>
              <h3 className="text-lg font-semibold mb-2">Add a Guest</h3>
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
                />
                <Button type="submit">
                  <PlusCircle className="h-4 w-4 mr-2" /> Add Guest
                </Button>
              </form>
            </div>
          </div>
        </CardContent>
        <CardFooter className="pt-6">
          <Button
            size="lg"
            className="w-full md:w-auto"
            disabled={players.length < 1}
          >
            Start Game
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
