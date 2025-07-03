import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { gameAPI } from "@/lib/api/service/game";
import { PageHeader } from "@/components/ui/page-header";
import { GameHistoryTable } from "@/components/games/games-history-table";
import { gameHistoryColumns as columns } from "@/components/games/games-history-columns";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Terminal } from "lucide-react";

export function GamePage() {
  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 10;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["gameHistory", page, pageSize, sorting],
    queryFn: () => gameAPI.getGameHistory(page, pageSize, sorting),
    placeholderData: (previousData) => previousData,
  });

  const renderContent = () => {
    if (isLoading) {
      return <Skeleton className="h-96 w-full" />;
    }

    if (isError) {
      const errorMessage =
        error instanceof Error ? error.message : "An unknown error occurred.";
      return (
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Game History</AlertTitle>
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      );
    }

    if (data?.data) {
      return (
        <div>
          <GameHistoryTable
            data={data.data.games}
            columns={columns}
            sorting={sorting}
            setSorting={setSorting}
          />
          <DataTablePagination
            pagination={data.data.pagination}
            setPage={setPage}
          />
        </div>
      );
    }

    return <p>No games found.</p>;
  };

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <PageHeader
        title="Game History"
        description="A record of all your completed games."
      />
      {renderContent()}
    </div>
  );
}
