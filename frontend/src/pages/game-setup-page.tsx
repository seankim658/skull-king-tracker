import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { DragEndEvent } from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useSubmit } from "@/hooks/use-submit";
import { gameAPI } from "@/lib/api/service/game";
import type { GamePlayerResponse } from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  GripVertical,
  Shuffle,
  NotebookPen,
  Rocket,
  AlertTriangle,
} from "lucide-react";
import { UserAvatar } from "@/components/ui/user-avatar";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/hooks/use-auth";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { InfoTooltip } from "@/components/ui/info-tooltip";

function SortablePlayerItem({ player }: { player: GamePlayerResponse }) {
  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id: player.game_player_id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      className="flex items-center justify-between bg-muted p-3 rounded-md"
    >
      <div className="flex items-center gap-4">
        <Button
          {...listeners}
          variant="ghost"
          size="icon"
          className="cursor-grab p-1"
        >
          <GripVertical className="h-5 w-5 text-muted-foreground" />
        </Button>
        <UserAvatar
          displayName={player.display_name}
          avatarUrl={player.avatar_url}
        />
        <span className="font-medium">{player.display_name}</span>
      </div>
    </div>
  );
}

export function GameSetupPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user: authenticatedUser } = useAuth();

  const [players, setPlayers] = useState<GamePlayerResponse[]>([]);
  const [scorekeeperId, setScorekeeperId] = useState<string>("");
  const [startingDealerId, setStartingDealerId] = useState<string>("");
  const [showScorekeeperWarning, setShowScorekeeperWarning] = useState(false);

  const { data: initialPlayers, isLoading: isLoadingPlayers } = useQuery({
    queryKey: ["gamePlayers", gameId],
    queryFn: async () => {
      const response = await gameAPI.getGamePlayers(gameId!);
      if (!response.success || !response.data) {
        throw new Error("Could not fetch players");
      }
      return response.data;
    },
    enabled: !!gameId,
  });

  const { data: gameDetails, isLoading: isLoadingDetails } = useQuery({
    queryKey: ["gameDetails", gameId],
    queryFn: async () => {
      const response = await gameAPI.getGameDetails(gameId!);
      if (!response.success || !response.data) {
        throw new Error("Could not fetch game details");
      }
      return response.data;
    },
    enabled: !!gameId,
  });

  useEffect(() => {
    if (initialPlayers) {
      setPlayers(initialPlayers);
      if (initialPlayers.length > 0) {
        setStartingDealerId(initialPlayers[0].game_player_id);
      }
    }
  }, [initialPlayers]);

  useEffect(() => {
    if (gameDetails?.current_scorekeeper_user_id) {
      setScorekeeperId(gameDetails.current_scorekeeper_user_id);
    }
  }, [gameDetails]);

  useEffect(() => {
    if (
      players.length > 0 &&
      !players.find((p) => p.game_player_id === startingDealerId)
    ) {
      setStartingDealerId(players[0].game_player_id);
    } else if (
      players.length > 0 &&
      players[0].game_player_id !== startingDealerId
    ) {
      const isDealerInList = players.some(
        (p) => p.game_player_id === startingDealerId,
      );
      if (!isDealerInList || players[0].game_player_id === startingDealerId) {
        setStartingDealerId(players[0].game_player_id);
      }
    }
  }, [players, startingDealerId]);

  useEffect(() => {
    if (
      authenticatedUser &&
      scorekeeperId &&
      scorekeeperId !== authenticatedUser.user_id
    ) {
      setShowScorekeeperWarning(true);
    } else {
      setShowScorekeeperWarning(false);
    }
  }, [scorekeeperId, authenticatedUser]);

  const { submit: saveSettings, isLoading: isSaving } = useSubmit(
    gameAPI.updateGameSettings,
    {
      actionVerb: "Saving settings",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["gameDetails", gameId] });
        queryClient.invalidateQueries({ queryKey: ["gamePlayers", gameId] });
      },
    },
  );

  const { submit: saveAndStart, isLoading: isStarting } = useSubmit(
    async (
      gameId: string,
      payload: {
        scorekeeper_user_id: string;
        ordered_player_ids: string[];
        starting_dealer_game_player_id: string;
      },
    ) => {
      const settingsResponse = await gameAPI.updateGameSettings(
        gameId,
        payload,
      );
      if (!settingsResponse.success) {
        throw new Error(
          settingsResponse.message || "Failed to save game settings",
        );
      }
      return gameAPI.startGame(gameId);
    },
    {
      actionVerb: "Saving and starting game",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["gameDetails", gameId] });
        queryClient.invalidateQueries({ queryKey: ["activeGames"] });
        navigate(`/game/${gameId}/scorecard`);
      },
    },
  );

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (over && active.id !== over.id) {
      setPlayers((items) => {
        const oldIndex = items.findIndex((p) => p.game_player_id === active.id);
        const newIndex = items.findIndex((p) => p.game_player_id === over.id);
        const newItems = arrayMove(items, oldIndex, newIndex);
        if (newItems.length > 0) {
          setStartingDealerId(newItems[0].game_player_id);
        }
        return newItems;
      });
    }
  }

  const handleShuffle = () => {
    const shuffled = [...players].sort(() => Math.random() - 0.5);
    setPlayers(shuffled);
    if (shuffled.length > 0) {
      setStartingDealerId(shuffled[0].game_player_id);
    }
  };

  const handleDealerChange = (newDealerId: string) => {
    setStartingDealerId(newDealerId);
    setPlayers((currentPlayers) => {
      const dealerIndex = currentPlayers.findIndex(
        (p) => p.game_player_id === newDealerId,
      );
      if (dealerIndex === -1) return currentPlayers;
      return [
        ...currentPlayers.slice(dealerIndex),
        ...currentPlayers.slice(0, dealerIndex),
      ];
    });
  };

  const handleSaveSettings = () => {
    if (
      !gameId ||
      !scorekeeperId ||
      !startingDealerId ||
      players.length === 0
    ) {
      toast.error("Cannot save settings: missing required information");
      return;
    }
    const orderedPlayerIds = players.map((p) => p.game_player_id);
    saveSettings(gameId, {
      scorekeeper_user_id: scorekeeperId,
      ordered_player_ids: orderedPlayerIds,
      starting_dealer_game_player_id: startingDealerId,
    });
  };

  const handleProceed = () => {
    if ((players?.length ?? 0) < 2) {
      toast.error("A game requires at least 2 players to proceed");
      return;
    }
    if (!gameId || !scorekeeperId || !startingDealerId) {
      toast.error("Cannot start game: missing required information");
      return;
    }

    saveAndStart(gameId, {
      scorekeeper_user_id: scorekeeperId,
      ordered_player_ids: players.map((p) => p.game_player_id),
      starting_dealer_game_player_id: startingDealerId,
    });
  };

  const registeredPlayers = players.filter((p) => p.user_id);
  const isCurrentUserScorekeeper =
    authenticatedUser?.user_id === gameDetails?.current_scorekeeper_user_id;

  if (isLoadingPlayers || isLoadingDetails) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Card className="max-w-2xl mx-auto">
          <CardHeader>
            <Skeleton className="h-8 w-3/5" />
            <Skeleton className="h-5 w-4/5" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 md:p-6">
      <Card className="max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle className="text-2xl">Game Setup</CardTitle>
          <CardDescription>
            Finalize the player order and settings before starting the game.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Seating Order Section */}
          <div>
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold">Seating Order</h3>
              <Button
                variant="outline"
                onClick={handleShuffle}
                disabled={isSaving}
                className="cursor-pointer"
              >
                <Shuffle className="h-4 w-4 mr-2" /> Randomize
              </Button>
            </div>
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={players.map((p) => p.game_player_id)}
                strategy={verticalListSortingStrategy}
              >
                <div className="space-y-2">
                  {players.map((player) => (
                    <SortablePlayerItem
                      key={player.game_player_id}
                      player={player}
                    />
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          </div>

          {/* Settings Section */}
          <div className="space-y-4">
            {/* Starting Dealer Section */}
            <div>
              <Label
                htmlFor="dealer-select"
                className="text-lg font-semibold block mb-2"
              >
                Starting Dealer
              </Label>
              <Select
                value={startingDealerId}
                onValueChange={handleDealerChange}
                disabled={isSaving}
              >
                <SelectTrigger id="dealer-select" className="w-full">
                  <SelectValue placeholder="Select starting dealer" />
                  <SelectContent>
                    {players.map((player) => (
                      <SelectItem
                        key={player.game_player_id}
                        value={player.game_player_id}
                      >
                        {player.display_name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </SelectTrigger>
              </Select>
            </div>

            {/* Scorekeeper Section */}
            <div>
              <div className="flex items-start gap-1">
                <Label
                  htmlFor="scorekeeper-select"
                  className="text-lg font-semibold block mb-2"
                >
                  Scorekeeper
                </Label>
                <InfoTooltip content="The scorekeeper is the only person who can edit the game state in any way (i.e. entering bids, entering tricks, completing the game, adding asterisks). You can transfer this role to another registered player, you cannot assign a guest player as scorekeeper." />
              </div>
              <Select
                value={scorekeeperId}
                onValueChange={setScorekeeperId}
                disabled={isSaving}
              >
                <SelectTrigger id="scorekeeper-select" className="w-full">
                  <SelectValue placeholder="Select a scorekeeper" />
                </SelectTrigger>
                <SelectContent>
                  {registeredPlayers.map((player) => (
                    <SelectItem key={player.user_id} value={player.user_id!}>
                      {player.display_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {showScorekeeperWarning && (
                <Alert variant="destructive" className="mt-3">
                  <AlertTriangle className="h-4 w-4" />
                  <AlertTitle>Warning: Transferring Control</AlertTitle>
                  <AlertDescription>
                    By changing the scorekeeper and saving, you will hand over
                    control of this game. You will no longer be able to make
                    changes.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          </div>
        </CardContent>

        <CardFooter className="flex flex-col sm:flex-row justify-between items-center gap-4 pt-6">
          <Button
            onClick={handleSaveSettings}
            disabled={isSaving || isStarting || !isCurrentUserScorekeeper}
            size="lg"
            className="w-full sm:w-auto cursor-pointer"
          >
            <NotebookPen className="h-4 w-4 mr-2" />
            {isSaving ? "Saving..." : "Save Settings"}
          </Button>
          <Button
            onClick={handleProceed}
            disabled={
              isSaving ||
              isStarting ||
              !isCurrentUserScorekeeper ||
              gameDetails?.status !== "pending"
            }
            size="lg"
            className="w-full sm:w-auto cursor-pointer"
          >
            <Rocket className="h-4 w-4 mr-2" />
            {gameDetails?.status === "pending"
              ? "Save and Start Game"
              : `Game is ${gameDetails?.status}`}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
