import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { sessionAPI } from "@/lib/api/service/session";
import { gameAPI } from "@/lib/api/service/game";
import { useConfirm } from "@/hooks/use-confirm";
import { useApi } from "@/hooks/use-api";
import { useSubmit } from "@/hooks/use-submit";
import type { ActiveSessionResponse, GameResponse } from "@/lib/api/types";

export function ActiveSessions() {
  const navigate = useNavigate();
  const confirm = useConfirm();

  const {
    data: activeSessions,
    isLoading: isLoadingSessions,
    request: fetchActiveSessions,
    setData: setActiveSessions,
  } = useApi(sessionAPI.getActiveSessionsForUser);

  useEffect(() => {
    fetchActiveSessions();
  }, [fetchActiveSessions]);

  const { submit: startGame, isLoading: isStartingGame } = useSubmit(
    gameAPI.createGame,
    {
      actionVerb: "Starting game",
      successMessage: "New game started",
      onSuccess: (data: GameResponse | undefined) => {
        if (data?.game_id) {
          navigate(`/game${data.game_id}/add-players`);
        }
      },
    },
  );

  const { submit: completeSession, isLoading: isCompletingSession } = useSubmit(
    sessionAPI.completeSession,
    {
      actionVerb: "Completing session",
      onSuccess: (_data, sessionId: string) => {
        setActiveSessions(
          (prev: ActiveSessionResponse[] | null) =>
            prev?.filter((s) => s.session_id !== sessionId) || [],
        );
      },
    },
  );

  const handleStartGameFromSession = async (sessionId: string) => {
    startGame({ session_id: sessionId });
  };

  const handleCompleteSession = async (
    sessionId: string,
    sessionName?: string,
  ) => {
    const isConfirmed = await confirm({
      title: "Are you sure?",
      description: `Do you want to makr the session "${sessionName || sessionId}" as completed? This action cannot be undone.`,
      confirmText: "Complete Session",
    });
    if (isConfirmed) {
      completeSession(sessionId);
    }
  };

  if (isLoadingSessions) {
    return (
      <section>
        <h2 className="text-2xl font-semibold mb-4">Your Active Sessions</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-6 w-3/4" />
                <Skeleton className="h-4 w-1/2 mt-1" />
              </CardHeader>
              <CardContent className="space-y-3">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-10 w-full" />
              </CardContent>
              <CardFooter className="flex flex-col sm:flex-row gap-2 pt-4">
                <Skeleton className="h-10 w-full sm:flex-1" />
                <Skeleton className="h-10 w-full sm:flex-1" />
              </CardFooter>
            </Card>
          ))}
        </div>
      </section>
    );
  }

  if (!activeSessions || activeSessions.length === 0) {
    return (
      <section>
        <h2 className="text-2xl font-semibold mb-4">Your Active Sessions</h2>
        <Card>
          <CardContent>
            <p className="text-center text-muted-foreground">
              No active sessions.
            </p>
          </CardContent>
        </Card>
      </section>
    );
  }

  return (
    <section>
      <h2 className="text-2xl font-semibold mb-6">Your Active Sessions</h2>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {activeSessions.map((session) => (
          <Card
            key={session.session_id}
            className="flex flex-col shadow-lg hover:shadow-xl transition-shadow duration-300"
          >
            <CardHeader>
              <CardTitle className="truncate text-xl">
                {session.session_name || "Unnamed Session"}
              </CardTitle>
              <CardDescription>
                ID: {session.session_id.substring(0, 8)}...
              </CardDescription>
            </CardHeader>
            <CardContent className="flex-grow space-y-2">
              <p className="text-sm">
                Status:{" "}
                <span className="font-medium capitalize text-primary">
                  {session.status}
                </span>
              </p>
              {session.has_active_game && (
                <p className="text-sm text-orange-500 font-semibold">
                  A game is currently in progress
                </p>
              )}
            </CardContent>
            <CardFooter className="flex flex-col sm:flex-row gap-3 pt-4 border-t mt-auto">
              <Button
                onClick={() => handleStartGameFromSession(session.session_id)}
                disabled={
                  session.has_active_game ||
                  isStartingGame ||
                  isCompletingSession
                }
                className="w-full sm:flex-1"
                variant="default"
              >
                {session.has_active_game ? "Game Active" : "Start New Game"}
              </Button>
              <Button
                variant="outline"
                onClick={() =>
                  handleCompleteSession(
                    session.session_id,
                    session.session_name,
                  )
                }
                disabled={isStartingGame || isCompletingSession}
                className="w-full sm:flex-1"
              >
                Complete Session
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>
    </section>
  );
}
