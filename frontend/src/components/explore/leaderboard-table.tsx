import { useQuery } from "@tanstack/react-query";
import { statsAPI } from "@/lib/api/service/stat";
import { DataTable } from "../ui/data-table";
import { leaderboardColumns } from "./leaderboard-columns";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { Terminal } from "lucide-react";
import { errorExtract } from "@/lib/utils";
import { Card } from "../ui/card";

export function LeaderboardTable() {
  const {
    data: leaderboardData,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["globalLeaderboard"],
    queryFn: async () => {
      const response = await statsAPI.getGlobalLeaderboard();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch leaderboard");
      }
      return response.data;
    },
    staleTime: 1000 * 60 * 60, // 1 hour
  });

  if (isError) {
    const errorMsg = errorExtract(error, "An unknown error occurred");
    return (
      <Card>
        <Alert variant="destructive" className="m-4">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Leaderboard</AlertTitle>
          <AlertDescription>{errorMsg}</AlertDescription>
        </Alert>
      </Card>
    );
  }

  return (
    <DataTable
      columns={leaderboardColumns}
      data={leaderboardData ?? []}
      isLoading={isLoading}
      loadingRowCount={10}
    />
  );
}
