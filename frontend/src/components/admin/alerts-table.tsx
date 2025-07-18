import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { alertAPI } from "@/lib/api/service/alert";
import type { SiteAlert, SiteAlertPayload } from "@/lib/api/types";
import { Button } from "../ui/button";
import { DataTable } from "../ui/data-table";
import { columns } from "./alert-columns";
import { AlertFormModal } from "./alert-form-modal";
import { useSubmit } from "@/hooks/use-submit";
import { useConfirm } from "@/hooks/use-confirm";

export function AlertsTable() {
  const queryClient = useQueryClient();
  const confirm = useConfirm();

  const [page, setPage] = useState(1);
  const [sorting, setSorting] = useState<SortingState>([]);
  const pageSize = 15;

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingAlert, setEditingAlert] = useState<SiteAlert | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["adminAlerts", page, pageSize, sorting],
    queryFn: () => alertAPI.getAdminAlerts({ page, pageSize, sorting }),
    placeholderData: (previousData) => previousData,
  });

  const { submit: createAlert, isLoading: isCreating } = useSubmit(
    alertAPI.createAlert,
    {
      actionVerb: "Creating alert",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["adminAlerts"] });
        setIsModalOpen(false);
      },
    },
  );

  const { submit: updateAlert, isLoading: isUpdating } = useSubmit(
    (payload: SiteAlertPayload) =>
      alertAPI.updateAlert(editingAlert!.alert_id, payload),
    {
      actionVerb: "Updating alert",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["adminAlerts"] });
        setIsModalOpen(false);
      },
    },
  );

  const { submit: deleteAlert, isLoading: isDeleting } = useSubmit(
    alertAPI.deleteAlert,
    {
      actionVerb: "Deleting alert",
      successMessage: "Alert deleted",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["adminAlerts"] });
      },
    },
  );

  const handleOpenCreateModal = () => {
    setEditingAlert(null);
    setIsModalOpen(true);
  };

  const handleOpenEditModal = (alert: SiteAlert) => {
    setEditingAlert(alert);
    setIsModalOpen(true);
  };

  const handleDeleteAlert = async (alert: SiteAlert) => {
    const isConfirmed = await confirm({
      title: "Delete Alert?",
      description: "This action cannot be undone.",
      confirmText: "Delete",
    });
    if (isConfirmed) {
      deleteAlert(alert.alert_id);
    }
  };

  const handleFormSubmit = (payload: SiteAlertPayload) => {
    if (editingAlert) {
      updateAlert(payload);
    } else {
      createAlert(payload);
    }
  };

  const alertColumns = columns({
    onEdit: handleOpenEditModal,
    onDelete: handleDeleteAlert,
  });

  return (
    <div>
      <div className="flex justify-end mb-4">
        <Button onClick={handleOpenCreateModal} className="cursor-pointer">
          Create New Alert
        </Button>
      </div>
      <DataTable
        columns={alertColumns}
        data={data?.data?.alerts ?? []}
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
      />
      <AlertFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleFormSubmit}
        isLoading={isCreating || isUpdating || isDeleting}
        alert={editingAlert}
      />
    </div>
  );
}
