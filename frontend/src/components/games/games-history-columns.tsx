import type { ColumnDef } from "@tanstack/react-table";
import type { GameHistoryItem } from "@/lib/api/types";
import { Button } from "../ui/button";
import { ArrowUpDown } from "lucide-react";

export const gameHistoryColumns: ColumnDef<GameHistoryItem>[] = [
  {
    accessorKey: "game_date",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Date
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <div>
        {new Date(row.original.game_date).toLocaleString("en-US", {
          dateStyle: "short",
          timeStyle: "short",
        })}
      </div>
    ),
  },
  {
    accessorKey: "session_name",
    header: "Session",
    cell: ({ row }) => row.original.session_name || "One-off Game",
  },
  {
    accessorKey: "finishing_position",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Position
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
  },
  {
    accessorKey: "total_points",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Total Points
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
  },
  {
    accessorKey: "rounds_hit",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Rounds Hit
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
  },
  {
    accessorKey: "zero_differential",
    header: "Zero Bid Net",
  },
  {
    accessorKey: "total_players",
    header: "Players",
  },
  {
    accessorKey: "scorekeeper_name",
    header: "Scorekeeper",
  },
];
