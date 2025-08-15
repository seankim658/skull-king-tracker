import { useState } from "react";
import { useSubmit } from "@/hooks/use-submit";
import { adminAPI } from "@/lib/api/service/admin";
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
  userToBan: { userId: string; displayName: string } | null;
}

export function BanUserModal({
  isOpen,
  onClose,
  userToBan,
}: BanUserModalProps) {
  const [reason, setReason] = useState("");
  const queryClient = useQueryClient();
  const { submit: banUser, isLoading } = useSubmit(adminAPI.banUser, {
    actionVerb: "Banning user",
    successMessage: `User ${userToBan?.displayName} has been banned.`,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["adminReports"] });
      queryClient.invalidateQueries({ queryKey: ["adminUsers"] });
      if (userToBan) {
        queryClient.invalidateQueries({
          queryKey: ["userProfile", userToBan.userId],
        });
      }
      onClose();
    },
  });

  const handleSubmit = () => {
    if (userToBan && reason.trim()) {
      banUser(userToBan.userId, { reason });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ban {userToBan?.displayName}</DialogTitle>
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
