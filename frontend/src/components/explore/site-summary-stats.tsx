import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { Users, Swords, CalendarDays, UserPlus } from "lucide-react";
import { statsAPI } from "@/lib/api/service/stat";
import type { SiteSummaryStatsResponse } from "@/lib/api/types";
import { toast } from "sonner";
import { errorExtract } from "@/lib/utils";
import { useApi } from "@/hooks/use-api";

const outerDivStyle = "grid gap-4 md:grid-cols-2 lg:grid-cols-4";
const summaryStatStyle = "text-2xl font-bold";
const iconStyle = "h-5 w-5 text-muted-foreground";
const cardHeaderStyle =
  "flex flex-row items-center justify-between space-y-0 pb-2";
const cardTitleStyle = "text-sm font-medium min-h-[50px]";

export function SiteSummaryStats() {
  const {
    data: summaryStats,
    isLoading,
    error,
    request: fetchSummary,
  } = useApi(statsAPI.getSiteSummaryStats);

  useEffect(() => {
    fetchSummary()
  }, [fetchSummary]);

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

  if (error || !summaryStats) {
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
      <Card>
        <CardHeader className={cardHeaderStyle}>
          <CardTitle className={cardTitleStyle}>
            Total Skull King Players
          </CardTitle>
          <Users className={iconStyle} />
        </CardHeader>
        <CardContent>
          <div className={summaryStatStyle}>{summaryStats.total_players}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className={cardHeaderStyle}>
          <CardTitle className={cardTitleStyle}>
            Sessions Played Last Month
          </CardTitle>
          <CalendarDays className={iconStyle} />
        </CardHeader>
        <CardContent>
          <div className={summaryStatStyle}>
            {summaryStats.sessions_this_month}
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className={cardHeaderStyle}>
          <CardTitle className={cardTitleStyle}>
            Games Played Last Month
          </CardTitle>
          <Swords className={iconStyle} />
        </CardHeader>
        <CardContent>
          <div className={summaryStatStyle}>
            {summaryStats.games_this_month}
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className={cardHeaderStyle}>
          <CardTitle className={cardTitleStyle}>
            New Players Last Month
          </CardTitle>
          <UserPlus className={iconStyle} />
        </CardHeader>
        <CardContent>
          <div className={summaryStatStyle}>
            {summaryStats.new_users_this_month}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
