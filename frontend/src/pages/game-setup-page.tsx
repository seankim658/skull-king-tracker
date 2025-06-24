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
import { useApi } from "@/hooks/use-api";
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
import { GripVertical, Shuffle, NotebookPen, Rocket } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getAvatarFallback, getFullAvatarURL } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";

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
        <Avatar className="h-10 w-10">
          <AvatarImage
            src={getFullAvatarURL(player.avatar_url)}
            alt={player.display_name}
          />
          <AvatarFallback>
            {getAvatarFallback(player.display_name)}
          </AvatarFallback>
        </Avatar>
        <span className="font-medium">{player.display_name}</span>
      </div>
    </div>
  );
}

export function GameSetupPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();

  const [players, setPlayers] = useState<GamePlayerResponse[]>([]);
  const [scorekeeperId, setScorekeeperId] = useState<string>("");

  const { data: initialPlayers, request: fetchPlayers } = useApi(
    gameAPI.getGamePlayers,
  );
  const { data: gameDetails, request: fetchGameDetails } = useApi(
    gameAPI.getGameDetails,
  );

  useEffect(() => {
    if (gameId) {
      fetchPlayers(gameId);
      fetchGameDetails(gameId);
    }
  }, [gameId, fetchPlayers, fetchGameDetails]);

  useEffect(() => {
    if (initialPlayers) setPlayers(initialPlayers);
  }, [initialPlayers]);

  useEffect(() => {
    if (gameDetails?.current_scorekeeper_user_id) {
      setScorekeeperId(gameDetails.current_scorekeeper_user_id);
    }
  }, [gameDetails]);

  const { submit: saveSettings, isLoading: isSaving } = useSubmit(
    gameAPI.updateGameSettings,
    {
      actionVerb: "Saving settings",
      onSuccess: () => {
        if (gameId) fetchGameDetails(gameId);
      },
    },
  );

  const { submit: startGame, isLoading: isStarting } = useSubmit(
    gameAPI.startGame,
    {
      actionVerb: "Starting game",
      onSuccess: () => navigate(`/game/${gameId}/scorecard`),
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
        return arrayMove(items, oldIndex, newIndex);
      });
    }
  }

  const handleShuffle = () => {
    setPlayers((currentPlayers) =>
      [...currentPlayers].sort(() => Math.random() - 0.5),
    );
  };

  const handleSaveSettings = () => {
    if (!gameId || !scorekeeperId || players.length === 0) {
      toast.error("Cannot save settings: missing required information");
      return;
    }
    const orderedPlayerIds = players.map((p) => p.game_player_id);
    saveSettings(gameId, {
      scorekeeper_user_id: scorekeeperId,
      ordered_player_ids: orderedPlayerIds,
    });
  };

  const registeredPlayers = players.filter((p) => p.user_id);

  if (!initialPlayers || !gameDetails) {
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

          {/* TODO Scorekeeper Section */}
          <div>
            <Label
              htmlFor="scorekeeper-select"
              className="text-lg font-semibold block mb-2"
            >
              Scorekeeper
            </Label>
            <Select
              value={scorekeeperId}
              onValueChange={setScorekeeperId}
              disabled={isSaving}
            >
              <SelectTrigger id="scorekeeper-select">
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
          </div>
        </CardContent>

        <CardFooter className="flex flex-col sm:flex-row justify-between items-center gap-4 pt-6">
          <Button
            onClick={handleSaveSettings}
            disabled={isSaving}
            size="lg"
            className="cursor-pointer"
          >
            <NotebookPen className="h-4 w-4 mr-2" />
            {isSaving ? "Saving..." : "Save Settings"}
          </Button>
          <Button
            onClick={() => gameId && startGame(gameId)}
            disabled={
              isSaving || isStarting || gameDetails.status !== "pending"
            }
            size="lg"
          >
            <Rocket className="h-4 w-4 mr-2 cursor-pointer" />
            {gameDetails.status === "pending"
              ? "Save and Start Game"
              : `Game is ${gameDetails.status}`}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
