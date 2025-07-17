import { useParams } from "react-router-dom";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "@/components/ui/card";
import {
  ScorecardHeader,
  ScorecardTable,
  ScorecardInputDrawer,
} from "@/components/scorecard";
import type {
  GamePlayerResponse,
  ScorecardResponse,
  SSEEvent,
} from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { PenLine, Terminal, Trophy } from "lucide-react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/hooks/use-auth";
import { gameAPI } from "@/lib/api/service/game";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useSubmit } from "@/hooks/use-submit";
import { AddAsteriskDialog } from "@/components/scorecard/add-asterisk-dialog";
import { GameSummaryModal } from "@/components/games/game-summary-modal";

export function GameScorecardPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const [isInputDrawerOpen, setIsInputDrawerOpen] = useState(false);
  const [isSummaryModalOpen, setIsSummaryModalOpen] = useState(false);

  const [isAsteriskDialogOpen, setIsAsteriskDialogOpen] = useState(false);
  const [selectedPlayerForAsterisk, setSelectedPlayerForAsterisk] =
    useState<GamePlayerResponse | null>(null);

  const {
    data: scorecardData,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["scorecard", gameId],
    queryFn: async () => {
      const response = await gameAPI.getScorecardState(gameId!);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch scorecard");
      }
      return response.data;
    },
    enabled: !!gameId,
  });

  const { submit: addAsterisk, isLoading: isAddingAterisk } = useSubmit(
    gameAPI.addAsterisk,
    {
      actionVerb: "Adding asterisk",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["scorecard", gameId] });
        setIsAsteriskDialogOpen(false);
        setSelectedPlayerForAsterisk(null);
      },
    },
  );

  const handleOpenAsteriskDialog = (player: GamePlayerResponse) => {
    setSelectedPlayerForAsterisk(player);
    setIsAsteriskDialogOpen(true);
  };

  const handleAddAsteriskSubmit = (playerId: string, reason: string) => {
    addAsterisk(gameId!, playerId, { reason });
  };

  useEffect(() => {
    if (!gameId) return;

    const eventsUrl = `${import.meta.env.VITE_SSE_BASE_URL}/notifications/events`;
    const eventSource = new EventSource(eventsUrl, { withCredentials: true });

    eventSource.onmessage = (event) => {
      try {
        const sseEvent: SSEEvent = JSON.parse(event.data);
        if (sseEvent.event === "scorecard_updated") {
          const updatedScorecard = sseEvent.payload as ScorecardResponse;
          if (updatedScorecard.game_id === gameId) {
            console.log("Received scorecard update via SSE:", updatedScorecard);
            queryClient.setQueryData(["scorecard", gameId], updatedScorecard);

            if (updatedScorecard.game_status === "completed" && user) {
              queryClient.invalidateQueries({
                queryKey: ["userProfile", user.user_id],
              });
              queryClient.invalidateQueries({
                queryKey: ["userAwardsStats", user.user_id],
              });
            }
          }
        }
      } catch (e) {
        console.error("Failed to parse SSE event data:", e);
      }
    };

    eventSource.onerror = (err) => {
      console.error("EventSource failed:", err);
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [gameId, queryClient, user]);

  if (isLoading) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-8 w-3/5" />
          </CardHeader>
          <CardHeader>
            <Skeleton className="h-64 w-full" />
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Scorecard</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!scorecardData) {
    return (
      <div className="container mx-auto p-4 md:p-6">No data available</div>
    );
  }

  const isScorekeeper =
    user?.user_id === scorecardData.current_scorekeeper_user_id;
  const currentRound = scorecardData.rounds.find(
    (r) => r.status === "bidding" || r.status === "playing",
  );

  const playersForInput =
    currentRound?.is_tie_breaker_round && scorecardData.overtime_player_ids
      ? scorecardData.players.filter((p) =>
          scorecardData.overtime_player_ids?.includes(p.game_player_id),
        )
      : scorecardData.players;

  const getButtonText = () => {
    if (!currentRound) return "View Scores";
    if (currentRound.status === "bidding") {
      return `Enter Bids for Round ${currentRound.round_number}`;
    }
    if (currentRound.status === "playing") {
      return `Enter Tricks for Round ${currentRound.round_number}`;
    }
    return "Game Over";
  };

  return (
    <div className="container mx-auto p-2 sm:p-4 md:p-6">
      <Card>
        <ScorecardHeader
          gameId={gameId!}
          sessionName={scorecardData.session_name}
          scorekeeperName={scorecardData.scorekeeper_name}
        />
        <CardContent className="p-0">
          <ScorecardTable
            players={scorecardData.players}
            rounds={scorecardData.rounds}
            asterisks={scorecardData.asterisks || []}
            isScorekeeper={isScorekeeper}
            onAddAsterisk={handleOpenAsteriskDialog}
          />
        </CardContent>

        <CardFooter className="flex-wrap justify-between pt-4 gap-2">
          {isScorekeeper &&
            scorecardData.game_status === "active" &&
            currentRound && (
              <Button
                size="lg"
                className="w-full sm:w-auto cursor-pointer"
                onClick={() => setIsInputDrawerOpen(true)}
              >
                <PenLine className="h-4 w-4 mr-2" />
                {getButtonText()}
              </Button>
            )}

          {scorecardData.game_status === "completed" && (
            <Button
              size="lg"
              className="w-full sm:w-auto cursor-pointer bg-amber-500 hover:bg-amber-600"
              onClick={() => setIsSummaryModalOpen(true)}
            >
              <Trophy className="h-4 w-4 mr-2" />
              View Game Summary
            </Button>
          )}
        </CardFooter>
      </Card>

      {currentRound && (
        <ScorecardInputDrawer
          isOpen={isInputDrawerOpen}
          onOpenChange={setIsInputDrawerOpen}
          currentRound={currentRound}
          players={playersForInput}
        />
      )}

      <AddAsteriskDialog
        player={selectedPlayerForAsterisk}
        isOpen={isAsteriskDialogOpen}
        onClose={() => setIsAsteriskDialogOpen(false)}
        onSubmit={handleAddAsteriskSubmit}
        isLoading={isAddingAterisk}
      />

      <GameSummaryModal
        gameId={gameId ?? null}
        isOpen={isSummaryModalOpen}
        onClose={() => setIsSummaryModalOpen(false)}
      />
    </div>
  );
}
