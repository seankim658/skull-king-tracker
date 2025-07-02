import { TableCell, TableRow } from "../ui/table";
import type { GamePlayerResponse } from "@/lib/api/types";

interface ScorecardTotalsRowProps {
  players: GamePlayerResponse[];
}

export function ScorecardTotalsRow({ players }: ScorecardTotalsRowProps) {
  return (
    <TableRow className="bg-muted hover:bg-muted font-bold text-base">
      <TableCell className="text-center">Total</TableCell>
      {players.map((player) => (
        <TableCell key={player.game_player_id} className="text-start text-lg">
          {player.final_score}
        </TableCell>
      ))}
    </TableRow>
  );
}
