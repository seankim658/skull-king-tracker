import { Button } from "./button";
import type { Pagination } from "@/lib/api/types";
import {
  ChevronsLeft,
  ChevronLeft,
  ChevronsRight,
  ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface DataTablePaginationProps {
  pagination: Pagination;
  setPage: (page: number) => void;
  className?: string;
}

export function DataTablePagination({
  pagination,
  setPage,
  className,
}: DataTablePaginationProps) {
  return (
    <div
      className={cn("flex items-center justify-end space-x-2 py-4", className)}
    >
      <div className="flex-1 text-sm text-muted-foreground">
        Page {pagination.current_page} of {pagination.total_pages}
      </div>
      <div className="flex items-center space-x-2">
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => setPage(1)}
          disabled={pagination.current_page <= 1}
        >
          <span className="sr-only">Go to first page</span>
          <ChevronsLeft className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => setPage(pagination.current_page - 1)}
          disabled={pagination.current_page <= 1}
        >
          <span className="sr-only">Go to previous page</span>
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => setPage(pagination.current_page + 1)}
          disabled={pagination.current_page >= pagination.total_pages}
        >
          <span className="sr-only">Go to next page</span>
          <ChevronRight className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          className="h-8 w-8 p-0"
          onClick={() => setPage(pagination.total_pages)}
          disabled={pagination.current_page >= pagination.total_pages}
        >
          <span className="sr-only">Go to last page</span>
          <ChevronsRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
