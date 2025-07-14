import type { ColumnDef } from "@tanstack/react-table";
import type { UserReport } from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";
import { ReportActions } from "./report-column-action";

type ReportsColumnsProps = {
  onBan: (report: UserReport) => void;
  onViewDetails: (report: UserReport) => void;
};

export const columns = ({
  onBan,
  onViewDetails,
}: ReportsColumnsProps): ColumnDef<UserReport>[] => [
  {
    accessorKey: "reporter_name",
    header: "Reporter",
    cell: ({ row }) => (
      <Button variant="link" asChild className="p-0 h-auto">
        <Link to={`/users/${row.original.reporter_user_id}`}>
          {row.original.reporter_name}
        </Link>
      </Button>
    ),
  },
  {
    accessorKey: "reported_name",
    header: "Reported",
    cell: ({ row }) => (
      <Button variant="link" asChild className="p-0 h-auto">
        <Link to={`/users/${row.original.reported_user_id}`}>
          {row.original.reported_name}
        </Link>
      </Button>
    ),
  },
  {
    accessorKey: "reason",
    header: "Reason",
    cell: ({ row }) => (
      <p className="truncate max-w-xs">{row.original.reason}</p>
    ),
  },
  {
    accessorKey: "created_at",
    header: "Date",
    cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString(),
  },
  {
    id: "actions",
    cell: ({ row }) => (
      <ReportActions
        report={row.original}
        onBan={onBan}
        onViewDetails={onViewDetails}
      />
    ),
  },
];
