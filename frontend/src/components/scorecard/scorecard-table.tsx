import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { getAvatarFallback, getFullAvatarURL } from "@/lib/utils";
import { ScorecardRow } from "./scorecard-row";
import { ScorecardTotalsRow } from "./scorecard-total-row";
import type { GamePlayerResponse, RoundScorecard } from "@/lib/api/types";

interface ScorecardTableProps {
  players: GamePlayerResponse[];
  rounds: RoundScorecard[];
}

export function ScorecardTable({ players, rounds }: ScorecardTableProps) {
  return (
    <div className="overflow-x-auto">
      <Table className="min-w-[600px]">
        <TableHeader>
          <TableRow className="bg-muted/30 hover:bg-muted/30">
            <TableHead className="w-[80px] font-bold text-center">
              Players
            </TableHead>
            {players.map((player) => (
              <TableHead
                key={player.game_player_id}
                className="text-center p-2"
              >
                <div className="flex flex-col items-center gap-1">
                  <Avatar className="h-10 w-10 border-2">
                    <AvatarImage src={getFullAvatarURL(player.avatar_url)} />
                    <AvatarFallback>
                      {getAvatarFallback(player.display_name)}
                    </AvatarFallback>
                  </Avatar>
                  <span className="font-semibold text-foreground truncate max-w-24">
                    {player.display_name}
                  </span>
                </div>
              </TableHead>
            ))}
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
