import type { ColumnDef } from "@tanstack/react-table";
import type { GlobalLeaderboardItem } from "@/lib/api/types";
import { Link } from "react-router-dom";
import { Button } from "../ui/button";

export const leaderboardColumns: ColumnDef<GlobalLeaderboardItem>[] = [
  {
    accessorKey: "rank",
    header: "Rank",
    cell: ({ row }) => (
      <div className="font-bold text-center w-8">{row.original.rank}</div>
    ),
  },
  {
    accessorKey: "player_name",
    header: "Player",
    cell: ({ row }) => (
      <Button asChild variant="link" className="p-0 h-auto">
        <Link to={`/users/${row.original.user_id}`}>
          {row.original.player_name}
        </Link>
      </Button>
    ),
  },
  {
    accessorKey: "games_played",
    header: "Games",
  },
  {
    accessorKey: "wins",
    header: "Wins",
  },
  {
    accessorKey: "total_points",
    header: "Total Points",
  },
  {
    accessorKey: "average_points",
    header: "Avg Points",
    cell: ({ row }) => row.original.average_points.toFixed(2),
  },
  {
    accessorKey: "average_finish_pos",
    header: "Avg Finish",
    cell: ({ row }) => row.original.average_finish_pos.toFixed(2),
  },
];
