import { Card, CardContent, CardHeader } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { Users, Swords, CalendarDays, UserPlus } from "lucide-react";
import { statsAPI } from "@/lib/api/service/stat";
import { useQuery } from "@tanstack/react-query";
import { StatCard } from "../ui/stat-card";

const outerDivStyle = "grid gap-4 grid-cols-2 lg:grid-cols-4";

export function SiteSummaryStats() {
  const {
    data: summaryStats,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["SiteSummaryStats"],
    queryFn: async () => {
      const response = await statsAPI.getSiteSummaryStats();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch size stats");
      }
      return response.data;
    },
    staleTime: 1000 * 60 * 15,
  });

  if (isLoading) {
    return (
      <div className={outerDivStyle}>
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <Skeleton className="h-5 w-3/5" />
              <Skeleton className="h-6 w-6 rounded-sm" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-1/2" />
              <Skeleton className="h-4 w-4/5 mt-1" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (isError || !summaryStats) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-center text-muted-foreground">
            Could not load site statistics.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className={outerDivStyle}>
      <StatCard
        title="Total Skull King Players"
        value={summaryStats.total_players}
        icon={<Users className="h-5 w-5 text-muted-foreground" />}
      />
      <StatCard
        title="Sessions Played Last Month"
        value={summaryStats.sessions_this_month}
        icon={<CalendarDays className="h-5 w-5 text-muted-foreground" />}
      />
      <StatCard
        title="Games Played Last Month"
        value={summaryStats.games_this_month}
        icon={<Swords className="h-5 w-5 text-muted-foreground" />}
      />
      <StatCard
        title="New Players Last Month"
        value={summaryStats.new_users_this_month}
        icon={<UserPlus className="h-5 w-5 text-muted-foreground" />}
      />
    </div>
  );
}
