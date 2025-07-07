import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { sessionAPI } from "@/lib/api/service/session";
import { PageHeader } from "@/components/ui/page-header";
import { DataTable } from "@/components/ui/data-table";
import { sessionHistoryColumns as columns } from "@/components/sessions/sessions-history-columns";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Terminal } from "lucide-react";
import { errorExtract } from "@/lib/utils";
import { useNavigate } from "react-router-dom";

export function SessionPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 15;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["sessionHistory", page, pageSize, sorting],
    queryFn: () => sessionAPI.getSessionHistory(page, pageSize, sorting),
    placeholderData: (previousData) => previousData,
  });

  const renderContent = () => {
    if (isError) {
      const errorMsg = errorExtract(error, "An unknown error occurred");
      return (
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Session History</AlertTitle>
          <AlertDescription>{errorMsg}</AlertDescription>
        </Alert>
      );
    }

    return (
      <DataTable
        columns={columns}
        data={data?.data?.sessions ?? []}
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
          navigate("/games", {
            state: {
              sessionId: row.original.session_id,
              sessionName: row.original.session_name || "Unnamed Session",
            },
          })
        }
      />
    );
  };

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <PageHeader
        title="Session History"
        description="A record of all your completed sessions."
      />
      {renderContent()}
    </div>
  );
}
