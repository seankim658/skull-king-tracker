import { TableCell, TableRow } from "../ui/table";
import { Badge } from "../ui/badge";
import { Swords, Shield, Spade, Star } from "lucide-react";
import { cn } from "@/lib/utils";
import type { GamePlayerResponse, RoundScorecard } from "@/lib/api/types";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";

interface ScorecardRowProps {
  round: RoundScorecard;
  players: GamePlayerResponse[];
}

export function ScorecardRow({ round, players }: ScorecardRowProps) {
  const isActiveRound = round.status !== "completed";

  const getPlayerScore = (playerId: string) => {
    return round.player_scores.find((s) => s.game_player_id === playerId);
  };

  return (
    <TableRow
      className={cn(
        isActiveRound &&
          "bg-primary/5 border-l-4 border-l-primary/80 dark:bg-primary/10",
      )}
    >
      <TableCell className="font-medium text-center sticky left-0 bg-background/95">
        <div className="flex flex-col items-center gap-1">
          <span className="text-xl">{round.round_number}</span>
          <Badge variant="secondary" className="px-1.5 py-0.5 text-xs">
            {round.round_number} cards
          </Badge>
        </div>
      </TableCell>

      {players.map((player) => {
        const score = getPlayerScore(player.game_player_id);
        const isDealer = false; // TODO : placeholder for now

        return (
          <TableCell
            key={player.game_player_id}
            className="p-2 pl-4 text-left relative min-w-[120px]"
          >
            {isDealer && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger className="absolute top-1.5 right-1.5">
                    <Spade className="h-4 w-4 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{player.display_name} is the dealer</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
            {/* TODO : Add dealer logic based on round.dealer_id */}
            <div className="flex flex-col items-start justify-start gap-1">
              <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Swords className="h-3.5 w-3.5" />
                <span>Bid: {score?.bid_amount ?? "-"}</span>
              </div>
              <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Shield className="h-3.5 w-3.5" />
                <span>Tricks: {score?.tricks_taken ?? "-"}</span>
              </div>
              <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Star className="h-3.5 w-3.5" />
                <span>Bonus: {score?.bonus_points ?? 0}</span>
              </div>
              <div className="font-bold text-lg mt-1">
                {score?.round_score ?? 0}
              </div>
            </div>
          </TableCell>
        );
      })}
    </TableRow>
  );
}
