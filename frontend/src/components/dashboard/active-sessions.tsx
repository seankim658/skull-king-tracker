import { useEffect, useState } from "react";
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
import { toast } from "sonner";
import { sessionAPI } from "@/lib/api/service/session";
import { gameAPI } from "@/lib/api/service/game";
import type { ActiveSessionResponse } from "@/lib/api/types";
import { errorExtract } from "@/lib/utils";

export function ActiveSessions() {
  const navigate = useNavigate();
  const [activeSessions, setActiveSessions] = useState<ActiveSessionResponse[]>(
    [],
  );
  const [isLoadingSessions, setIsLoadingSessions] = useState(true);

  const fetchActiveSessions = async () => {
    setIsLoadingSessions(true);
    try {
      const response = await sessionAPI.getActiveSessionsForUser();
      if (response.success && response.data) {
        setActiveSessions(response.data);
      } else {
        toast.error(response.message || "Failed to load active sessions");
        setActiveSessions([]);
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Could not fetch active sessions");
      toast.error(errMsg);
      console.error(errMsg);
      setActiveSessions([]);
    } finally {
      setIsLoadingSessions(false);
    }
  };

  useEffect(() => {
    fetchActiveSessions();
  }, []);

  const handleStartGameFromSession = async (
    sessionId: string,
    sessionName?: string,
  ) => {
    const toastId = toast.loading(
      `Starting new game in session "${sessionName || sessionId}"...`,
    );
    try {
      const response = await gameAPI.createGame({ session_id: sessionId });
      if (response.success && response.data?.game_id) {
        toast.success("New game started", { id: toastId });
        // TODO : navigate to add players page
      } else {
        toast.error(response.message || "Failed to start new game in session", {
          id: toastId,
        });
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Failed to start new game in session");
      toast.error(errMsg);
      console.error(errMsg);
    }
  };

  const handleCompleteSession = async (
    sessionId: string,
    sessionName?: string,
  ) => {
    // TODO : make pretty complete dialog later
    if (
      !confirm(
        `Are you sure you want to mark the session "${sessionName || sessionId}" as completed? This action cannot be undone.`,
      )
    ) {
      return;
    }
    const toastId = toast.loading(
      `Completing session "${sessionName || sessionId}"...`,
    );
    try {
      const response = await sessionAPI.completeSession(sessionId);
      if (response.success) {
        toast.success(
          `Session "${sessionName || sessionId}" marked as completed`,
          { id: toastId },
        );
        setActiveSessions((prevSessions) =>
          prevSessions.filter((s) => s.session_id !== sessionId),
        );
      } else {
        toast.error(response.message || "Failed to complete session", {
          id: toastId,
        });
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Could not complete session");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
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

  if (activeSessions.length === 0) {
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
                onClick={() =>
                  handleStartGameFromSession(
                    session.session_id,
                    session.session_name,
                  )
                }
                disabled={session.has_active_game}
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
