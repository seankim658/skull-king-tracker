import type { ColumnDef } from "@tanstack/react-table";
import type { SessionHistoryItem } from "@/lib/api/types";
import { SortableHeader } from "../ui/data-table-sortable-header";

export const sessionHistoryColumns: ColumnDef<SessionHistoryItem>[] = [
  {
    accessorKey: "session_name",
    header: "Session Name",
    cell: ({ row }) => row.original.session_name || "Unnamed Session",
  },
  {
    accessorKey: "date_completed",
    header: ({ column }) => (
      <SortableHeader column={column} title="Date Completed" />
    ),
    cell: ({ row }) =>
      new Date(row.original.date_completed).toLocaleDateString(),
  },
  {
    accessorKey: "number_of_games",
    header: ({ column }) => (
      <SortableHeader column={column} title="Games Played" />
    ),
  },
  {
    accessorKey: "your_wins",
    header: ({ column }) => (
      <SortableHeader column={column} title="Your Wins" />
    ),
  },
  {
    accessorKey: "win_percentage",
    header: ({ column }) => <SortableHeader column={column} title="Win %" />,
    cell: ({ row }) => `${row.original.win_percentage.toFixed(2)}%`,
  },
  {
    accessorKey: "average_finishing_position",
    header: ({ column }) => (
      <SortableHeader column={column} title="Avg. Position" />
    ),
    cell: ({ row }) => row.original.average_finishing_position.toFixed(2),
  },
  {
    accessorKey: "session_creator",
    header: "Creator",
  },
];
