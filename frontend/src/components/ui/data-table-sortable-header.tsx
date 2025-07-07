import { Button } from "../ui/button";
import { ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import type { Column } from "@tanstack/react-table";
import { InfoTooltip } from "./info-tooltip";
import type { ReactNode } from "react";

interface SortableHeaderProps<T> {
  column: Column<T>;
  title: string;
  tooltipContent?: ReactNode;
}

export const SortableHeader = <T,>({
  column,
  title,
  tooltipContent,
}: SortableHeaderProps<T>) => {
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
      <div className="relative inline-flex items-center">
        <div className="mr-2">{title}</div>
        {tooltipContent && (
          <InfoTooltip content={tooltipContent} position="inline" />
        )}
      </div>
      {renderIcon()}
    </Button>
  );
};
