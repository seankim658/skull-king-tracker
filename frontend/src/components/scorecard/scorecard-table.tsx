import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { UserAvatar } from "../ui/user-avatar";
import { ScorecardRow } from "./scorecard-row";
import { ScorecardTotalsRow } from "./scorecard-total-row";
import type {
  GamePlayerResponse,
  RoundScorecard,
  PlayerGameAsterisk,
} from "@/lib/api/types";
import { Asterisk, PlusCircle } from "lucide-react";
import { Button } from "../ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";
import { cn } from "@/lib/utils";

interface ScorecardTableProps {
  players: GamePlayerResponse[];
  rounds: RoundScorecard[];
  asterisks: PlayerGameAsterisk[];
  isScorekeeper: boolean;
  onAddAsterisk: (player: GamePlayerResponse) => void;
}

const asteriskPositions = [
  "top-0 -right-1", // 1. Top-right
  "top-1/2 -right-2 -translate-y-1/2", // 2. Middle-right
  "bottom-0 -right-1", // 3. Bottom-right
  "top-0 -left-1", // 4. Top-left
  "top-1/2 -left-2 -translate-y-1/2", // 5. Middle-left
  "bottom-0 -left-1", // 6. Bottom-left
  "left-1/2 -top-2 -translate-x-1/2", // 7. Top-middle
  "left-1/2 -bottom-2 -translate-x-1/2", // 8. Bottom-middle
];

export function ScorecardTable({
  players,
  rounds,
  asterisks,
  isScorekeeper,
  onAddAsterisk,
}: ScorecardTableProps) {
  return (
    <div className="overflow-x-auto">
      <Table className="min-w-[600px]">
        <TableHeader>
          <TableRow className="bg-muted/30 hover:bg-muted/30">
            <TableHead className="w-[80px] font-bold text-center sticky left-0 bg-background z-20">
              Players
            </TableHead>
            {players.map((player) => {
              const playerAsterisks =
                asterisks?.filter(
                  (a) => a.game_player_id === player.game_player_id,
                ) || [];

              return (
                <TableHead
                  key={player.game_player_id}
                  className="text-center p-2"
                >
                  <div className="flex flex-col items-center gap-2">
                    <div className="relative h-10 w-10">
                      <UserAvatar
                        displayName={player.display_name}
                        avatarUrl={player.avatar_url}
                        className="h-10 w-10 border-2"
                      />
                      {playerAsterisks.map((asterisk, index) => (
                        <TooltipProvider key={asterisk.player_game_asterisk_id}>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Asterisk
                                className={cn(
                                  "absolute h-3.5 w-3.5 text-destructive bg-card rounded-sm p-px",
                                  asteriskPositions[
                                    index % asteriskPositions.length
                                  ],
                                )}
                              />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p>{asterisk.reason || "General tomfoolery"}</p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      ))}
                    </div>

                    <div className="flex items-center gap-1">
                      <span className="font-semibold text-foreground truncate max-w-24">
                        {player.display_name}
                      </span>
                      {isScorekeeper && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-5 w-5 rounded-full cursor-pointer"
                          onClick={() => onAddAsterisk(player)}
                        >
                          <PlusCircle className="h-4 w-4 text-muted-foreground" />
                        </Button>
                      )}
                    </div>
                  </div>
                </TableHead>
              );
            })}
          </TableRow>
        </TableHeader>
        <TableBody>
          {rounds.map((round) => (
            <ScorecardRow
              key={round.round_number}
              round={round}
              players={players}
            />
          ))}
          <ScorecardTotalsRow players={players} />
        </TableBody>
      </Table>
    </div>
  );
}
