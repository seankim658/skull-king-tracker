import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Label } from "../ui/label";
import type { UserReport } from "@/lib/api/types";
import { Link } from "react-router-dom";
import { Badge } from "../ui/badge";

interface ReportDetailsModalProps {
  isOpen: boolean;
  onClose: () => void;
  report: UserReport | null;
}

export function ReportDetailsModal({
  isOpen,
  onClose,
  report,
}: ReportDetailsModalProps) {
  if (!report) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Report Details</DialogTitle>
          <DialogDescription>
            Full details for the report submitted on{" "}
            {new Date(report.created_at).toLocaleString()}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="flex justify-between items-center">
            <div>
              <Label>Reporter</Label>
              <Button variant="link" asChild className="p-0 h-auto block">
                <Link to={`/users/${report.reporter_user_id}`}>
                  {report.reporter_name}
                </Link>
              </Button>
            </div>
            <div>
              <Label>Reported User</Label>
              <Button variant="link" asChild className="p-0 h-auto block">
                <Link to={`/users/${report.reported_user_id}`}>
                  {report.reported_name}
                </Link>
              </Button>
            </div>
            <div>
              <Label>Status</Label>
              <div className="mt-1">
                <Badge variant="secondary" className="capitalize">
                  {report.status}
                </Badge>
              </div>
            </div>
          </div>
          <div>
            <Label>Reason</Label>
            <p className="mt-1 text-sm p-3 bg-muted rounded-md whitespace-pre-wrap">
              {report.reason}
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
