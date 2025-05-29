import { useAuth } from "@/hooks/use-auth";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import { ActiveSessions } from "@/components/dashboard/active-sessions";

export function DashboardPage() {
  const { user } = useAuth();

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <ActiveSessions />
    </div>
  );
}
