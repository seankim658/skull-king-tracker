import { useState } from "react";
import { useSubmit } from "@/hooks/use-submit";
import { adminAPI } from "@/lib/api/service/admin";
import type { UserReport } from "@/lib/api/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useQueryClient } from "@tanstack/react-query";

interface BanUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  userReport: UserReport | null;
}

export function BanUserModal({
  isOpen,
  onClose,
  userReport,
}: BanUserModalProps) {
  const [reason, setReason] = useState("");
  const queryClient = useQueryClient();
  const { submit: banUser, isLoading } = useSubmit(adminAPI.banUser, {
    actionVerb: "Banning user",
    successMessage: `User ${userReport?.reported_name} has been banned.`,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["adminReports"] });
      if (userReport) {
        queryClient.invalidateQueries({
          queryKey: ["userProfile", userReport.reported_user_id],
        });
      }
      onClose();
    },
  });

  const handleSubmit = () => {
    if (userReport && reason.trim()) {
      banUser(userReport.reported_user_id, { reason });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ban {userReport?.reported_name}</DialogTitle>
        </DialogHeader>
        <div className="py-4">
          <Label htmlFor="ban-reason">Reason for Ban</Label>
          <Textarea
            id="ban-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="mt-2"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isLoading || !reason.trim()}
            variant="destructive"
          >
            Confirm Ban
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
