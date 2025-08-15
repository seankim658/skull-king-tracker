import type { ColumnDef } from "@tanstack/react-table";
import type { AdminUserView } from "@/lib/api/types";
import { UserAvatar } from "../ui/user-avatar";
import { Badge } from "../ui/badge";
import { UserActions } from "./user-column-actions";

type UserColumnsProps = {
  onBan: (user: AdminUserView) => void;
};

export const columns = ({
  onBan,
}: UserColumnsProps): ColumnDef<AdminUserView>[] => [
  {
    id: "index",
    header: "#",
    cell: ({ row }) => <span>{row.index + 1}</span>,
  },
  {
    accessorKey: "display_name",
    header: "User",
    cell: ({ row }) => {
      const user = row.original;
      const displayName = user.display_name || user.username;
      return (
        <div className="flex items-center gap-2">
          <UserAvatar
            userId={user.user_id}
            displayName={displayName}
            avatarUrl={user.avatar_url}
            updatedAt={user.updated_at}
            className="h-8 w-8"
          />
          <div className="flex flex-col">
            <span className="font-medium">{displayName}</span>
            <span className="text-xs text-muted-foreground">
              @{user.username}
            </span>
          </div>
        </div>
      );
    },
  },
  {
    accessorKey: "email",
    header: "Email",
  },
  {
    accessorKey: "role",
    header: "Role",
    cell: ({ row }) => (
      <Badge variant="outline" className="capitalize">
        {row.original.role}
      </Badge>
    ),
  },
  {
    accessorKey: "is_banned",
    header: "Status",
    cell: ({ row }) =>
      row.original.is_banned ? (
        <Badge variant="destructive">Banned</Badge>
      ) : (
        <Badge variant="secondary">Active</Badge>
      ),
  },
  {
    accessorKey: "ban_reason",
    header: "Ban Reason",
  },
  {
    accessorKey: "last_login_at",
    header: "Last Login",
    cell: ({ row }) =>
      row.original.last_login_at
        ? new Date(row.original.last_login_at).toLocaleString()
        : "Never",
  },
  {
    accessorKey: "created_at",
    header: "Joined",
    cell: ({ row }) => new Date(row.original.created_at).toLocaleDateString(),
  },
  {
    id: "actions",
    cell: ({ row }) => <UserActions user={row.original} onBan={onBan} />,
  },
];
