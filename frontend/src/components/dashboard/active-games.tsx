import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { gameAPI } from "@/lib/api/service/game";
import { Button } from "../ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { UserAvatar } from "../ui/user-avatar";
import { NotebookPen, Calendar, Users, Rocket } from "lucide-react";
import { InfoTooltip } from "../ui/info-tooltip";

export function ActiveGames() {
  const navigate = useNavigate();
  const { data: activeGames, isLoading } = useQuery({
    queryKey: ["activeGames"],
    queryFn: async () => {
      const response = await gameAPI.getActiveGames();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch active games");
      }
      return response.data;
    },
    staleTime: 1000 * 60,
  });

  if (isLoading) {
    return (
      <section>
        <h2 className="text-2xl font-semibold mb-4">Your Active Games</h2>
        <div className="flex flex-col gap-4">
          <Skeleton className="h-40 w-full" />
        </div>
      </section>
    );
  }

  if (!activeGames || activeGames.length === 0) {
    return null;
  }

  return (
    <section>
      <div className="mb-4">
        <div className="relative inline-block">
          <h2 className="text-2xl font-semibold mr-2">Your Active Games</h2>
          <InfoTooltip
            content={
              <>
                This section shows all games that are currently in progress.
                <br />
                <br />
                If you are the <strong>scorekeeper</strong>, you can tap to
                resume scoring. Otherwise, you can view the live scorecard.
              </>
            }
          />
        </div>
      </div>

      <div className="grid gap-6">
        {activeGames.map((game) => (
          <Card key={game.game_id} className="flex flex-col">
            <CardHeader>
              <CardTitle className="flex justify-between items-center">
                <span>{game.session_name || "One-off Game"}</span>
                <span className="text-sm font-medium text-muted-foreground">
                  Round {game.current_round}
                </span>
              </CardTitle>
              <div className="text-xs text-muted-foreground flex flex-wrap gap-x-4 gap-y-1 pt-1">
                <span className="flex items-center gap-1.5">
                  <NotebookPen className="h-3.5 w-3.5" />
                  Scorekeeper: {game.scorekeeper_name}
                </span>
                <span className="flex items-center gap-1.5">
                  <Calendar className="h-3.5 w-3.5" />
                  Started:{" "}
                  {new Date(game.created_at).toLocaleString("en-US", {
                    dateStyle: "short",
                    timeStyle: "short",
                  })}
                </span>
              </div>
            </CardHeader>
            <CardContent className="flex-grow">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-muted-foreground" />
                <h4 className="font-medium text-sm">Players</h4>
              </div>
              <div className="flex flex-wrap items-center gap-3 mt-2">
                {game.players.map((player, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <UserAvatar
                      displayName={player.display_name}
                      avatarUrl={player.avatar_url}
                      className="h-8 w-8 border"
                    />
                    <span className="text-sm font-medium">
                      {player.display_name}
                    </span>
                  </div>
                ))}
              </div>
            </CardContent>
            <div className="p-4 pt-2">
              <Button
                onClick={() => navigate(`/game/${game.game_id}/scorecard`)}
                className="w-full cursor-pointer"
              >
                <Rocket className="h-4 w-4 mr-2" />
                {game.is_scorekeeper ? "Resume Scoring" : "View Scorecard"}
              </Button>
            </div>
          </Card>
        ))}
      </div>
    </section>
  );
}
