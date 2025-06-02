import type { UserStats } from "@/lib/api/types";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Gamepad2, Trophy, Percent, Medal } from "lucide-react";

const cardHeaderStyling = "flex flex-row items-center justify-between";

interface ProfileStatsSummaryProps {
  stats?: UserStats;
  username: string;
}

export function ProfileStatsSummary({
  stats,
  username,
}: ProfileStatsSummaryProps) {
  if (!stats) {
    return (
      <div>
        <h2 className="text-2xl font-semibold mb-4">
          Statistics for {username}
        </h2>
        <Card>
          <CardContent className="pt-6">
            <p className="text-muted-foreground">
              Statistics are private or no game data available.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-semibold mb-4">Statistics for {username}</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className={cardHeaderStyling}>
            <CardTitle className="text-sm font-medium">
              Total Games Played
            </CardTitle>
            <Gamepad2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_games_played}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium">Total Wins</CardTitle>
            <Trophy className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_wins}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium">
              Win Percentage
            </CardTitle>
            <Percent className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.win_percentage.toFixed(2)}%
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
