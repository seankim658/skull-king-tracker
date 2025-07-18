import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { alertAPI } from "@/lib/api/service/alert";
import { Alert, AlertDescription, AlertTitle } from "./alert";
import { Megaphone, X } from "lucide-react";
import { Button } from "./button";

export function SiteAlertBanner() {
  const [dismissedAlerts, setDismissedAlerts] = useState<string[]>([]);
  const { data: alerts } = useQuery({
    queryKey: ["activeAlerts"],
    queryFn: async () => {
      const response = await alertAPI.getActiveAlerts();
      return response.data ?? [];
    },
    staleTime: 1000 * 60 * 5,
  });

  if (!alerts || alerts.length === 0) {
    return null;
  }

  const activeAlerts = alerts.filter(
    (alert) => !dismissedAlerts.includes(alert.alert_id),
  );

  if (activeAlerts.length === 0) {
    return null;
  }

  return (
    <div className="container mx-auto px-4 md:px-6 pt-4 space-y-2">
      {activeAlerts.map((alert) => (
        <Alert key={alert.alert_id}>
          <Megaphone className="h-4 w-4" />
          <div className="flex-grow">
            <AlertTitle>{alert.title}</AlertTitle>
            <AlertDescription>{alert.body}</AlertDescription>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={() =>
              setDismissedAlerts((prev) => [...prev, alert.alert_id])
            }
          >
            <X className="h-4 w-4" />
          </Button>
        </Alert>
      ))}
    </div>
  );
}
