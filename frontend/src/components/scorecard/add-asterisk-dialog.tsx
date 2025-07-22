import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "../ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "../ui/select";
import { Button } from "../ui/button";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import type { GamePlayerResponse } from "@/lib/api/types";

interface AddAsteriskDialogProps {
  player: GamePlayerResponse | null;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (playerId: string, reason: string) => void;
  isLoading: boolean;
}

const ASTERISK_REASONS = [
  "Dealing faux pas",
  "Illegal play",
  "Table talk",
  "Being sus",
  "Being generally dumb",
  "Other",
];

export function AddAsteriskDialog({
  player,
  isOpen,
  onClose,
  onSubmit,
  isLoading,
}: AddAsteriskDialogProps) {
  const [selectedReason, setSelectedReason] = useState("");
  const [customReason, setCustomReason] = useState("");

  useEffect(() => {
    if (isOpen) {
      setSelectedReason("");
      setCustomReason("");
    }
  }, [isOpen]);

  const handleSubmit = () => {
    if (player) {
      const finalReason =
        selectedReason === "Other" ? customReason : selectedReason;
      onSubmit(player.game_player_id, finalReason);
    }
  };

  if (!player) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Asterisk for {player.display_name}</DialogTitle>
          <DialogDescription>
            Record a misplay or other notable event.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4 space-y-4">
          <div className="space-y-2">
            <Label htmlFor="reason-select">Reason</Label>
            <Select
              value={selectedReason}
              onValueChange={setSelectedReason}
              disabled={isLoading}
            >
              <SelectTrigger id="reason-select">
                <SelectValue placeholder="Select a reason..." />
              </SelectTrigger>
              <SelectContent>
                {ASTERISK_REASONS.map((reason) => (
                  <SelectItem key={reason} value={reason}>
                    {reason}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {selectedReason === "Other" && (
            <div className="space-y-2">
              <Label htmlFor="custom-reason">Custom Reason</Label>
              <Input
                id="custom-reason"
                value={customReason}
                onChange={(e) => setCustomReason(e.target.value)}
                placeholder="Describe the intent..."
                disabled={isLoading}
              />
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={onClose}
              disabled={isLoading}
              className="cursor-pointer"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isLoading || selectedReason === ""}
              className="cursor-pointer"
            >
              {isLoading ? "Adding..." : "Add Asterisk"}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  );
}
