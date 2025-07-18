import type { ColumnDef } from "@tanstack/react-table";
import type { SiteAlert } from "@/lib/api/types";
import { Badge } from "@/components/ui/badge";
import { AlertActions } from "./alert-actions";

type AlertColumnsProps = {
  onEdit: (alert: SiteAlert) => void;
  onDelete: (alert: SiteAlert) => void;
};

export const columns = ({
  onEdit,
  onDelete,
}: AlertColumnsProps): ColumnDef<SiteAlert>[] => [
  {
    accessorKey: "title",
    header: "Title",
  },
  {
    accessorKey: "start_time",
    header: "Starts",
    cell: ({ row }) => new Date(row.original.start_time).toLocaleString(),
  },
  {
    accessorKey: "end_time",
    header: "Ends",
    cell: ({ row }) => new Date(row.original.end_time).toLocaleString(),
  },
  {
    accessorKey: "is_active",
    header: "Status",
    cell: ({ row }) => {
      const now = new Date();
      const start = new Date(row.original.start_time);
      const end = new Date(row.original.end_time);
      const isCurrentlyRunning =
        row.original.is_active && now >= start && now <= end;

      if (isCurrentlyRunning) {
        return (
          <Badge className="bg-green-600 hover:bg-green-700">Active</Badge>
        );
      }
      if (row.original.is_active && now < start) {
        return <Badge variant="secondary">Scheduled</Badge>;
      }
      return <Badge variant="outline">Inactive</Badge>;
    },
  },
  {
    accessorKey: "creator_name",
    header: "Created By",
  },
  {
    id: "actions",
    cell: ({ row }) => (
      <AlertActions alert={row.original} onEdit={onEdit} onDelete={onDelete} />
    ),
  },
];
