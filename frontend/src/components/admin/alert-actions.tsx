import type { SiteAlert } from "@/lib/api/types";
import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuSeparator,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { MoreHorizontal, Trash2, Pencil } from "lucide-react";

export function AlertActions({
  alert,
  onEdit,
  onDelete,
}: {
  alert: SiteAlert;
  onEdit: (alert: SiteAlert) => void;
  onDelete: (alert: SiteAlert) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0 cursor-pointer">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => onEdit(alert)}
          className="cursor-pointer"
        >
          <Pencil className="mr-2 h-4 w-4" />
          Edit Alert
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => onDelete(alert)}
          className="text-destructive cursor-pointer"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete Alert
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
