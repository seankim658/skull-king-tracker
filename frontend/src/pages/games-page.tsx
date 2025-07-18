import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { gameAPI } from "@/lib/api/service/game";
import { PageHeader } from "@/components/ui/page-header";
import { DataTable } from "@/components/ui/data-table";
import { gameHistoryColumns as columns } from "@/components/games/games-history-columns";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Terminal } from "lucide-react";
import { errorExtract } from "@/lib/utils";
import { useNavigate, useLocation, Link } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export function GamePage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { sessionId, sessionName } = location.state || {
    sessionId: null,
    sessionName: null,
  };

  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 10;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["gameHistory", page, pageSize, sorting, sessionId],
    queryFn: () => gameAPI.getGameHistory(page, pageSize, sorting, sessionId),
    placeholderData: (previousData) => previousData,
  });

  const renderContent = () => {
    if (isError) {
      const errorMsg = errorExtract(error, "An unknown error occurred");
      return (
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Game History</AlertTitle>
          <AlertDescription>{errorMsg}</AlertDescription>
        </Alert>
      );
    }

    return (
      <DataTable
        columns={columns}
        data={data?.data?.games ?? []}
        sorting={sorting}
        setSorting={setSorting}
        pagination={
          data?.data?.pagination ?? {
            current_page: 1,
            total_pages: 1,
            page_size: pageSize,
            total_count: 0,
          }
        }
        setPage={setPage}
        isLoading={isLoading}
        onRowClick={(row) =>
          navigate(`/game/${row.original.game_id}/scorecard`, {
            state: {
              from: "/games",
              label: sessionId ? sessionName : "Game History",
              previousState: location.state,
            },
          })
        }
      />
    );
  };

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      {sessionId ? (
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink asChild>
                <Link to="/sessions">Sessions</Link>
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{sessionName ? `${sessionName} Games` : "Session Games"}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      ) : (
        <PageHeader
          title="Game History"
          description="A record of all your completed games."
        />
      )}
      {renderContent()}
    </div>
  );
}
