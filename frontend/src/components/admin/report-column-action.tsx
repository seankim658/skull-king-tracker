import type { UserReport } from "@/lib/api/types";
import { useConfirm } from "@/hooks/use-confirm";
import { useQueryClient } from "@tanstack/react-query";
import { useSubmit } from "@/hooks/use-submit";
import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuSeparator,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { MoreHorizontal, Ban, ShieldOff, FileText } from "lucide-react";
import { adminAPI } from "@/lib/api/service/admin";

export function ReportActions({
  report,
  onBan,
  onViewDetails,
}: {
  report: UserReport;
  onBan: (report: UserReport) => void;
  onViewDetails: (report: UserReport) => void;
}) {
  const confirm = useConfirm();
  const queryClient = useQueryClient();

  const { submit: dismissReport } = useSubmit(adminAPI.updateReportStatus, {
    actionVerb: "Dismissing report",
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["adminReports"] }),
    successMessage: "Report dismissed",
  });

  const handleDismiss = async () => {
    const isConfirmed = await confirm({
      title: "Dismiss Report?",
      description:
        "Are you sure you want to dismiss this report? This action cannot be undone.",
      confirmText: "Dismiss",
    });
    if (isConfirmed) {
      dismissReport(report.report_id, { status: "dismissed" });
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => onViewDetails(report)}
          className="cursor-pointer"
        >
          <FileText className="mr-2 h-4 w-4" />
          View Details
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => onBan(report)}
          className="text-destructive cursor-pointer"
        >
          <Ban className="mr-2 h-4 w-4" />
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleDismiss} className="cursor-pointer">
          <ShieldOff className="mr-2 h-4 w-4" />
          Dismiss Report
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
