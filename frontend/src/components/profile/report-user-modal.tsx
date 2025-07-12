import { useState } from "react";
import { useSubmit } from "@/hooks/use-submit";
import { userAPI } from "@/lib/api/service/user";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "../ui/dialog";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";

interface ReportUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  reportedUser: { userId: string; displayName: string };
}

export function ReportUserModal({
  isOpen,
  onClose,
  reportedUser,
}: ReportUserModalProps) {
  const [reason, setReason] = useState("");
  const { submit: submitReport, isLoading } = useSubmit(userAPI.reportUser, {
    actionVerb: "Submitting report",
    successMessage: "Report submitted successfully",
    onSuccess: onClose,
  });

  const handleSubmit = () => {
    if (reason.trim()) {
      submitReport(reportedUser.userId, { reason });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Report {reportedUser.displayName}</DialogTitle>
          <DialogDescription>
            Please provide a reason for reporting this user. Your report will be
            reviewed by an administrator.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          <Label htmlFor="reason">Reason for reporting</Label>
          <Textarea
            id="reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Describe the issue..."
            className="mt-2"
            rows={4}
          />
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isLoading}
            className="cursor-pointer"
          >
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading || !reason.trim()}>
            {isLoading ? "Submitting..." : "Submit Report"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
