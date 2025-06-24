import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useApi } from "@/hooks/use-api";
import { sessionAPI } from "@/lib/api/service/session";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "../ui/dialog";
import { Skeleton } from "../ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import { cn } from "@/lib/utils";
import {
  List,
  Terminal,
  Trophy,
  Swords,
  Rocket,
  CheckCircle,
  NotebookPen,
} from "lucide-react";

interface SessionDetailModalProps {
  sessionId: string | null;
  isOpen: boolean;
  onClose: () => void;
}

export function SessionDetailsModal({
  sessionId,
  isOpen,
  onClose,
}: SessionDetailModalProps) {
  const navigate = useNavigate();
  const {
    data: sessionDetails,
    isLoading,
    error,
    request: fetchDetails,
  } = useApi(sessionAPI.getSessionDetails);

  useEffect(() => {
    if (sessionId && isOpen) {
      fetchDetails(sessionId);
    }
  }, [sessionId, isOpen, fetchDetails]);

  const handleNavigateToGame = (gameId: string, status: string) => {
    if (status === "pending") {
      navigate(`/game/${gameId}/add-players`);
    } else {
      navigate(`/game/${gameId}/scorecard`);
    }
    onClose();
  };

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <Skeleton className="h-20" />
            <Skeleton className="h-20" />
          </div>
          <Skeleton className="h-px w-full" />
          <div className="space-y-3">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        </div>
      );
    }

    if (error || !sessionDetails) {
      return (
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Session</AlertTitle>
          <AlertDescription>
            {error || "Could not load the session details"}
          </AlertDescription>
        </Alert>
      );
    }

    return (
      <div className="space-y-4">
        {/* Stats Summary */}
        <div className="grid grid-cols-2 gap-4">
          <div className="p-4 border rounded-lg">
            <div className="flex items-center gap-3 mb-1">
              <Swords className="h-5 w-5 text-muted-foreground" />
              <h4 className="font-semibold">Total Games</h4>
            </div>
            <p className="text-3xl font-bold">
              {sessionDetails.user_summary.total_games}
            </p>
          </div>
          <div className="p-4 border rounded-lg">
            <div className="flex items-center gap-3 mb-1">
              <Trophy className="h-5 w-5 text-muted-foreground" />
              <h4 className="font-semibold">Your Wins</h4>
            </div>
            <p className="text-3xl font-bold">
              {sessionDetails.user_summary.wins}
            </p>
          </div>
        </div>

        {/* Games List */}
        <div>
          <h4 className="font-semibold text-lg mb-2 flex items-center">
            <List className="mr-2 h-5 w-5" /> Games
          </h4>
          <div className="space-y-2 mad-h-[40vh] overflow-y-auto pr-2">
            {sessionDetails.games.length === 0 ? (
              <p className="text-muted-foreground text-center py-4">
                No games have been played in this session yet.
              </p>
            ) : (
              sessionDetails.games.map((game) => (
                <div
                  key={game.game_id}
                  className="flex items-center justify-between p-3 bg-muted/50 rounded-md"
                >
                  <div>
                    <p className="font-medium flex items-center">
                      Game{" "}
                      <Badge
                        variant={
                          game.status === "active" ? "destructive" : "secondary"
                        }
                        className={cn("ml-2 capitalize", {
                          "animate-pulse": game.status === "active",
                        })}
                      >
                        {game.status}
                      </Badge>
                    </p>
                    {game.scorekeeper_name && (
                      <p className="text-sm text-muted-foreground flex items-center mt-1">
                        <NotebookPen className="h-4 w-4 mr-1.5 text-blue-500" />
                        Scorekeeper: {game.scorekeeper_name}
                      </p>
                    )}
                    {game.status === "completed" && game.winning_player && (
                      <p className="text-sm text-muted-foreground flex items-center mt-1">
                        <CheckCircle className="h-4 w-4 mr-1.5 text-green-500" />
                        Winner: {game.winning_player}
                      </p>
                    )}
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() =>
                      handleNavigateToGame(game.game_id, game.status)
                    }
                    className="cursor-pointer"
                    disabled={game.status === "pending" && !game.is_scorekeeper}
                  >
                    <Rocket className="h-4 w-4 mr-2" />
                    {game.status === "pending"
                      ? "Setup"
                      : game.status === "active"
                        ? "Resume"
                        : "Scorecard"}
                  </Button>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className="truncate text-2xl">
            {sessionDetails?.session_name || "Session Details"}
          </DialogTitle>
          <DialogDescription>A summary of this session.</DialogDescription>
        </DialogHeader>
        <div className="py-2">{renderContent()}</div>
      </DialogContent>
    </Dialog>
  );
}
