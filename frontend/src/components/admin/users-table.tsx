import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { adminAPI } from "@/lib/api/service/admin";
import type { AdminUserView } from "@/lib/api/types";
import { DataTable } from "@/components/ui/data-table";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { columns } from "@/components/admin/user-columns";
import { BanUserModal } from "@/components/admin/ban-user-modal";
import { errorExtract } from "@/lib/utils";
import { Terminal } from "lucide-react";

export function UsersTable() {
  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 20;

  const [isBanModalOpen, setBanModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<AdminUserView | null>(null);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["adminUsers", page, pageSize, sorting],
    queryFn: () => adminAPI.getUsers({ page, pageSize }),
    placeholderData: (previousData) => previousData,
  });

  const handleBanUser = (user: AdminUserView) => {
    setSelectedUser(user);
    setBanModalOpen(true);
  };

  if (isError) {
    const errorMsg = errorExtract(error, "An unknown error occurred");
    return (
      <Alert variant="destructive">
        <Terminal className="h-4 w-4" />
        <AlertTitle>Error Loading Users</AlertTitle>
        <AlertDescription>{errorMsg}</AlertDescription>
      </Alert>
    );
  }

  const userColumns = columns({
    onBan: handleBanUser,
  });

  const userToBan = selectedUser
    ? {
        userId: selectedUser.user_id,
        displayName: selectedUser.display_name || selectedUser.username,
      }
    : null;

  return (
    <>
      <DataTable
        columns={userColumns}
        data={data?.data?.users ?? []}
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
        userToBan={userToBan}
      />
    </>
  );
}
