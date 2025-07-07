import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { sessionAPI } from "@/lib/api/service/session";
import { gameAPI } from "@/lib/api/service/game";
import { useConfirm } from "@/hooks/use-confirm";
import { useSubmit } from "@/hooks/use-submit";
import type { ActiveSessionResponse, GameResponse } from "@/lib/api/types";
import { SessionDetailsModal } from "./session-details-modal";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { InfoTooltip } from "../ui/info-tooltip";

export function ActiveSessions() {
  const navigate = useNavigate();
  const confirm = useConfirm();
  const queryClient = useQueryClient();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(
    null,
  );

  const { data: activeSessions, isLoading: isLoadingSessions } = useQuery({
    queryKey: ["activeSessions"],
    queryFn: async () => {
      const response = await sessionAPI.getActiveSessionsForUser();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch active sessions");
      }
      return response.data;
    },
  });

  const { submit: startGame, isLoading: isStartingGame } = useSubmit(
    gameAPI.createGame,
    {
      actionVerb: "Starting game",
      successMessage: "New game started",
      onSuccess: (data: GameResponse | undefined) => {
        queryClient.invalidateQueries({ queryKey: ["activeSessions"] });
        queryClient.invalidateQueries({ queryKey: ["activeGames"] });
        if (data?.game_id) {
          navigate(`/game/${data.game_id}/add-players`);
        }
      },
    },
  );

  const { submit: completeSession, isLoading: isCompletingSession } = useSubmit(
    sessionAPI.completeSession,
    {
      actionVerb: "Completing session",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["activeSessions"] });
        queryClient.invalidateQueries({ queryKey: ["sessionHistory"] });
      },
    },
  );

  const handleStartNewGameClick = async (session: ActiveSessionResponse) => {
    if (session.has_active_game || session.has_pending_game) {
      const isConfirmed = await confirm({
        title: "In-progress game exists",
        description:
          "This session already has an active/pending game. Are you sure you want to start a new one?",
        confirmText: "Start New Game",
      });
      if (!isConfirmed) return;
    }
    startGame({ session_id: session.session_id });
  };

  const handleCompleteSession = async (
    sessionId: string,
    sessionName?: string,
  ) => {
    const isConfirmed = await confirm({
      title: "Are you sure?",
      description: `Do you want to mark the session "${sessionName || sessionId}" as completed? This action cannot be undone.`,
      confirmText: "Complete Session",
    });
    if (isConfirmed) {
      completeSession(sessionId);
    }
  };

  const handleCardClick = (sessionId: string) => {
    setSelectedSessionId(sessionId);
    setIsModalOpen(true);
  };

  if (isLoadingSessions) {
    return (
      <section>
        <h2 className="text-2xl font-semibold mb-4">Your Active Sessions</h2>
        <div className="flex flex-col gap-4">
          {[...Array(2)].map((_, i) => (
            <Skeleton key={i} className="h-24 w-full" />
          ))}
        </div>
      </section>
    );
  }

  if (!activeSessions || activeSessions.length === 0) {
    return (
      <section>
        <h2 className="text-2xl font-semibold mb-4">Your Active Sessions</h2>
        <Card className="flex items-center justify-center p-12">
          <p className="text-center text-muted-foreground">
            No active sessions. Start a new one from the sidebar!
          </p>
        </Card>
      </section>
    );
  }

  return (
    <section>
      <div className="mb-4">
        <div className="relative inline-block">
          <h2 className="text-2xl font-semibold mr-2">Your Active Sessions</h2>
          <InfoTooltip
            content={
              <>
                A <strong>Session</strong> is a collection of multiple games,
                perfect for tracking a group of games. You can start a new game
                within a session here, or tap a session to see more details.
              </>
            }
          />
        </div>
      </div>

      <div className="flex flex-col gap-4">
        {activeSessions.map((session) => (
          <Card
            key={session.session_id}
            className="flex flex-col sm:flex-row items-center justify-center p-4 gap-4 transition-all hover:shadow-md cursor-pointer"
            onClick={() => handleCardClick(session.session_id)}
          >
            <div className="flex flex-grow items-center gap-4 w-full">
              {session.has_active_game && (
                <div className="w-3 h-3">
                  <span className="relative flex h-3 w-3">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                  </span>
                </div>
              )}
              <div className="flex-grow">
                <h3 className="text-lg font-bold">
                  {session.session_name || "Unnamed Session"}
                </h3>
                <p className="text-sm text-muted-foreground">
                  Created by{" "}
                  <span className="font-medium text-foreground">
                    {session.creator_name || "Unknown"}
                  </span>
                  {" on "}
                  {new Date(session.created_at).toLocaleDateString("en-US", {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                    hour: "numeric",
                    minute: "2-digit",
                  })}
                </p>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row flex-shrink-0 gap-2 w-full sm:w-auto">
              <Button
                onClick={(e) => {
                  e.stopPropagation();
                  handleStartNewGameClick(session);
                }}
                disabled={
                  session.has_active_game ||
                  session.has_pending_game ||
                  isStartingGame ||
                  isCompletingSession
                }
                className="w-full sm:flex-1 cursor-pointer"
                variant="default"
              >
                Start New Game
              </Button>
              <Button
                variant="outline"
                onClick={(e) => {
                  e.stopPropagation();
                  handleCompleteSession(
                    session.session_id,
                    session.session_name,
                  );
                }}
                disabled={isStartingGame || isCompletingSession}
                className="w-full sm:flex-1 cursor-pointer"
              >
                Complete Session
              </Button>
            </div>
          </Card>
        ))}
      </div>
      <SessionDetailsModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        sessionId={selectedSessionId}
      />
    </section>
  );
}
