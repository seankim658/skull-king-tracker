import { Button } from "../ui/button";
import { ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import type { Column } from "@tanstack/react-table";

export const SortableHeader = <T,>({
  column,
  title,
}: {
  column: Column<T>;
  title: string;
}) => {
  const sortDirection = column.getIsSorted();

  const renderIcon = () => {
    if (sortDirection === "asc") {
      return <ArrowUp className="ml-2 h-4 w-4" />;
    }
    if (sortDirection === "desc") {
      return <ArrowDown className="ml-2 h-4 w-4" />;
    }
    return <ArrowUpDown className="ml-2 h-4 w-4" />;
  };

  return (
    <Button
      variant="ghost"
      onClick={() => column.toggleSorting(sortDirection === "asc")}
    >
      {title}
      {renderIcon()}
    </Button>
  );
};
