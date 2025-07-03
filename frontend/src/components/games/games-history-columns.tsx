import type { ColumnDef } from "@tanstack/react-table";
import type { GameHistoryItem } from "@/lib/api/types";
import { SortableHeader } from "../ui/data-table-sortable-header";

export const gameHistoryColumns: ColumnDef<GameHistoryItem>[] = [
  {
    accessorKey: "game_date",
    header: ({ column }) => <SortableHeader column={column} title="Date" />,
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
    header: ({ column }) => <SortableHeader column={column} title="Position" />,
  },
  {
    accessorKey: "total_points",
    header: ({ column }) => (
      <SortableHeader column={column} title="Total Points" />
    ),
  },
  {
    accessorKey: "rounds_hit",
    header: ({ column }) => (
      <SortableHeader column={column} title="Rounds Hit" />
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
