import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { UserAvatar } from "../ui/user-avatar";
import { NotebookPen, Calendar, Users, Rocket } from "lucide-react";
import type { ActiveGameResponse } from "@/lib/api/types";

interface GameCardProps {
  game: ActiveGameResponse;
  type: "active" | "pending";
}

export function GameCard({ game, type }: GameCardProps) {
  const navigate = useNavigate();

  const handleNavigation = () => {
    if (type === "pending") {
      navigate(`/game/${game.game_id}/add-players`);
    } else {
      navigate(`/game/${game.game_id}/scorecard`);
    }
  };

  const getButtonText = () => {
    if (type === "pending") {
      return game.is_scorekeeper ? "Resume Setup" : "View Lobby";
    }
    return game.is_scorekeeper ? "Resume Scoring" : "View Scorecard";
  };

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle className="flex justify-between items-center">
          <span>{game.session_name || "One-off Game"}</span>
          {type === "active" && (
            <span className="text-sm font-medium text-muted-foreground">
              Round {game.current_round}
            </span>
          )}
        </CardTitle>
        <div className="text-xs text-muted-foreground flex flex-wrap gap-x-4 gap-y-1 pt-1">
          <span className="flex items-center gap-1.5">
            <NotebookPen className="h-3.5 w-3.5" />
            Scorekeeper: {game.scorekeeper_name}
          </span>
          <span className="flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            Created:{" "}
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
                userId={player.user_id}
                displayName={player.display_name}
                avatarUrl={player.avatar_url}
                updatedAt={player.updated_at}
                className="h-8 w-8 border"
              />
              <span className="text-sm font-medium">{player.display_name}</span>
            </div>
          ))}
        </div>
      </CardContent>
      <div className="p-4 pt-2">
        <Button
          onClick={handleNavigation}
          className="w-full cursor-pointer"
          disabled={type === "pending" && !game.is_scorekeeper}
        >
          <Rocket className="h-4 w-4 mr-2" />
          {getButtonText()}
        </Button>
      </div>
    </Card>
  );
}
