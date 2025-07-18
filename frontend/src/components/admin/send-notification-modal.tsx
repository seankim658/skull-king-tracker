import { useState } from "react";
import { useSubmit } from "@/hooks/use-submit";
import { adminAPI } from "@/lib/api/service/admin";
import type { UserSearchItem } from "@/lib/api/types";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "../ui/dialog";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import { Checkbox } from "../ui/checkbox";
import { UserSearch } from "../explore/user-search";
import { Badge } from "../ui/badge";
import { X } from "lucide-react";
import { DialogDescription } from "@radix-ui/react-dialog";

interface SendNotificationModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SendNotificationModal({
  isOpen,
  onClose,
}: SendNotificationModalProps) {
  const [message, setMessage] = useState("");
  const [isBroadcast, setIsBroadcast] = useState(false);
  const [recipients, setRecipients] = useState<UserSearchItem[]>([]);

  const { submit: sendNotification, isLoading } = useSubmit(
    adminAPI.sendAdminNotification,
    {
      actionVerb: "Sending notification",
      onSuccess: () => {
        setMessage("");
        setRecipients([]);
        setIsBroadcast(false);
        onClose();
      },
    },
  );

  const handleSubmit = () => {
    const user_ids = isBroadcast ? [] : recipients.map((r) => r.user_id);
    sendNotification({ message, is_broadcast: isBroadcast, user_ids });
  };

  const handleAddRecipient = (user: UserSearchItem) => {
    if (!recipients.some((r) => r.user_id === user.user_id)) {
      setRecipients((prev) => [...prev, user]);
    }
  };

  const handleRemoveRecipient = (userId: string) => {
    setRecipients((prev) => prev.filter((r) => r.user_id !== userId));
  };

  const isSubmitDisabled =
    isLoading || !message.trim() || (!isBroadcast && recipients.length === 0);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Send Admin Notification</DialogTitle>
          <DialogDescription>
            Compose a message to send to all users or specific users.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="flex items-center space-x-2">
            <Checkbox
              id="broadcast"
              checked={isBroadcast}
              onCheckedChange={(checked) => setIsBroadcast(!!checked)}
              disabled={isLoading}
            />
            <Label htmlFor="broadcast">Send to all users (broadcast)</Label>
          </div>
          {!isBroadcast && (
            <div className="space-y-2">
              <Label>Recipients</Label>
              <UserSearch onUserSelect={handleAddRecipient} />
              <div className="flex flex-wrap gap-2 pt-2 min-h-6">
                {recipients.map((user) => (
                  <Badge key={user.user_id} variant="secondary">
                    {user.display_name || user.username}
                    <Button
                      onClick={() => handleRemoveRecipient(user.user_id)}
                      className="ml-1.5 rounded-full hover:bg-background/50 p-0.5"
                      aria-label={`Remove ${user.display_name || user.username}`}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </Badge>
                ))}
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="message">Message</Label>
            <Textarea
              id="message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              disabled={isLoading}
              rows={5}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={isLoading} className="cursor-pointer">
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isSubmitDisabled} className="cursor-pointer">
            {isLoading ? "Sending..." : "Send Notification"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
