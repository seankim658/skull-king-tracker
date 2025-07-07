import type { UserStats } from "@/lib/api/types";
import { Gamepad2, Trophy, Percent, Medal } from "lucide-react";
import { StatCard } from "../ui/stat-card";
import { Card, CardContent } from "../ui/card";

const iconStyling = "h-4 w-4 text-muted-foreground";

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
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Games Played"
          value={stats.total_games_played}
          icon={<Gamepad2 className={iconStyling} />}
        />
        <StatCard
          title="Total Wins"
          value={stats.total_wins}
          icon={<Trophy className={iconStyling} />}
        />
        <StatCard
          title="Top 3 Finishes"
          value={stats.top_3_finishes}
          icon={<Medal className={iconStyling} />}
          tooltip="Only counts games with 4 or more players."
        />
        <StatCard
          title="Win Percentage"
          value={stats.win_percentage}
          icon={<Percent className={iconStyling} />}
        />
      </div>
    </div>
  );
}
