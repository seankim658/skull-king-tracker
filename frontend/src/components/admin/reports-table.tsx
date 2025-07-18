import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { adminAPI, type ReportFilters } from "@/lib/api/service/admin";
import type { UserReport } from "@/lib/api/types";
import { DataTable } from "@/components/ui/data-table";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { columns } from "@/components/admin/report-columns";
import { BanUserModal } from "@/components/admin/ban-user-modal";
import { errorExtract } from "@/lib/utils";
import { Terminal } from "lucide-react";
import { ReportDetailsModal } from "@/components/admin/report-details-modal";

export function ReportsTable() {
  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 20;

  const [isBanModalOpen, setBanModalOpen] = useState(false);
  const [isDetailModalOpen, setDetailsModalOpen] = useState(false);
  const [selectedReport, setSelectedReport] = useState<UserReport | null>(null);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["adminReports", page, pageSize, sorting],
    queryFn: () => {
      const filters: ReportFilters = {
        page,
        pageSize,
        status: "pending",
        sorting,
      };
      return adminAPI.getReports(filters);
    },
    placeholderData: (previousData) => previousData,
  });

  const handleBanUser = (report: UserReport) => {
    setSelectedReport(report);
    setBanModalOpen(true);
  };

  const handleViewDetails = (report: UserReport) => {
    setSelectedReport(report);
    setDetailsModalOpen(true);
  };

  if (isError) {
    const errorMsg = errorExtract(error, "An unknown error occurred");
    return (
      <Alert variant="destructive">
        <Terminal className="h-4 w-4" />
        <AlertTitle>Error Loading Reports</AlertTitle>
        <AlertDescription>{errorMsg}</AlertDescription>
      </Alert>
    );
  }

  const reportColumns = columns({
    onBan: handleBanUser,
    onViewDetails: handleViewDetails,
  });

  return (
    <>
      <DataTable
        columns={reportColumns}
        data={data?.data?.reports ?? []}
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
        loadingRowCount={pageSize}
      />
      <BanUserModal
        isOpen={isBanModalOpen}
        onClose={() => setBanModalOpen(false)}
        userReport={selectedReport}
      />
      <ReportDetailsModal
        isOpen={isDetailModalOpen}
        onClose={() => setDetailsModalOpen(false)}
        report={selectedReport}
      />
    </>
  );
}
