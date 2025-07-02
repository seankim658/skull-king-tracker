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
  ScorecardResponse,
  SSEEvent,
} from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { PenLine, Terminal } from "lucide-react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/hooks/use-auth";
import { gameAPI } from "@/lib/api/service/game";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function GameScorecardPage() {
  const { gameId } = useParams<{ gameId: string }>();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const [isInputDrawerOpen, setIsInputDrawerOpen] = useState(false);

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
  }, [gameId, queryClient]);

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
          />
        </CardContent>

        {isScorekeeper &&
          scorecardData.game_status === "active" &&
          currentRound && (
            <CardFooter className="pt-2">
              <Button
                size="lg"
                className="w-full sm:w-auto cursor-pointer"
                onClick={() => setIsInputDrawerOpen(true)}
              >
                <PenLine className="h-4 w-4 mr-2" />
                {getButtonText()}
              </Button>
            </CardFooter>
          )}
      </Card>

      {currentRound && (
        <ScorecardInputDrawer
          isOpen={isInputDrawerOpen}
          onOpenChange={setIsInputDrawerOpen}
          currentRound={currentRound}
          players={scorecardData.players}
        />
      )}
    </div>
  );
}
