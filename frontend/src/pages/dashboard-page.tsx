import { ActiveSessions } from "@/components/dashboard/active-sessions";
import { Separator } from "@/components/ui/separator";
import { ActiveGames } from "@/components/dashboard/active-games";
import { PendingGames } from "@/components/dashboard/pending-games";

export function DashboardPage() {

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <ActiveSessions />
      <Separator />
      <ActiveGames />
      <PendingGames />
    </div>
  );
}
