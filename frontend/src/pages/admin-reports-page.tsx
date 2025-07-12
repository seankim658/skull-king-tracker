import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { adminAPI, type ReportFilters } from "@/lib/api/service/admin";
import type { UserReport } from "@/lib/api/types";
import { PageHeader } from "@/components/ui/page-header";
import { DataTable } from "@/components/ui/data-table";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { columns } from "@/components/admin/report-columns";
import { BanUserModal } from "@/components/admin/ban-user-modal";
import { SendNotificationModal } from "@/components/admin/send-notification-modal";
import { errorExtract } from "@/lib/utils";
import { Terminal, Send } from "lucide-react";

export function AdminReportsPage() {
  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 20;

  const [isBanModalOpen, setBanModalOpen] = useState(false);
  const [isNotificationModalOpen, setNotificationModalOpen] = useState(false);
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

  const renderContent = () => {
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
    const reportColumns = columns({ onBan: handleBanUser });

    return (
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
    );
  };

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <div className="flex justify-between items-center">
        <PageHeader
          title="Admin Panel: User Reports"
          description="Review and act on user-submitted reports."
        />
        <Button onClick={() => setNotificationModalOpen(true)}>
          <Send className="mr-2 h-4 w-4" />
          Send Notification
        </Button>
      </div>

      {renderContent()}

      {/* Modals */}
      <BanUserModal
        isOpen={isBanModalOpen}
        onClose={() => setBanModalOpen(false)}
        userReport={selectedReport}
      />
      <SendNotificationModal
        isOpen={isNotificationModalOpen}
        onClose={() => setNotificationModalOpen(false)}
      />
    </div>
  );
}
