import { useState, useEffect } from "react";
import type { SiteAlert, SiteAlertPayload } from "@/lib/api/types";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "../ui/dialog";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";
import { Switch } from "../ui/switch";

interface AlertFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (payload: SiteAlertPayload) => void;
  isLoading: boolean;
  alert?: SiteAlert | null;
}

export function AlertFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  alert,
}: AlertFormModalProps) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [isActive, setIsActive] = useState(true);

  useEffect(() => {
    if (isOpen && alert) {
      setTitle(alert.title);
      setBody(alert.body);
      setStartTime(new Date(alert.start_time).toISOString().slice(0, 16));
      setEndTime(new Date(alert.end_time).toISOString().slice(0, 16));
      setIsActive(alert.is_active);
    } else if (isOpen) {
      setTitle("");
      setBody("");
      setStartTime("");
      setEndTime("");
      setIsActive(true);
    }
  }, [isOpen, alert]);

  const handleSubmit = () => {
    const payload: SiteAlertPayload = {
      title,
      body,
      start_time: new Date(startTime),
      end_time: new Date(endTime),
      is_active: isActive,
    };
    onSubmit(payload);
  };

  const isFormValid = title && body && startTime && endTime;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{alert ? "Edit Alert" : "Create New Alert"}</DialogTitle>
          <DialogDescription>
            This alert will be shown to all users on the site between the start
            and end times.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="title" className="text-right">
              Title
            </Label>
            <Input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-start gap-4">
            <Label htmlFor="body" className="text-right pt-2">
              Body
            </Label>
            <Textarea
              id="body"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="col-span-3"
              rows={4}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="start-time" className="text-right">
              Start Time
            </Label>
            <Input
              id="start-time"
              type="datetime-local"
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="end-time" className="text-right">
              End Time
            </Label>
            <Input
              id="end-time"
              type="datetime-local"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="is-active" className="text-right">
              Active
            </Label>
            <Switch
              id="is-active"
              checked={isActive}
              onCheckedChange={setIsActive}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading || !isFormValid}>
            {isLoading ? "Saving..." : "Save Alert"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
