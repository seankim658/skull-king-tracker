import { useQuery } from "@tanstack/react-query";
import { gameAPI } from "@/lib/api/service/game";
import { Skeleton } from "../ui/skeleton";
import { Card } from "../ui/card";
import { GameCard } from "./game-card";
import { DashboardSectionHeader } from "./dashboard-section-header";

export function PendingGames() {
  const { data: pendingGames, isLoading } = useQuery({
    queryKey: ["pendingGames"],
    queryFn: async () => {
      const response = await gameAPI.getPendingGames();
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch pending games");
      }
      return response.data;
    },
    staleTime: 1000 * 60,
  });

  const tooltipContent = (
    <>
      This section shows all games you are participating in that are pending the
      completion of the setup process.
      <br />
      <br />
      If you are the <strong>scorekeeper</strong>, you can tap to resume game
      setup.
    </>
  );

  return (
    <section>
      <DashboardSectionHeader
        title="Pending Games"
        tooltipContent={tooltipContent}
      />
      {isLoading ? (
        <div className="flex flex-col gap-4">
          <Skeleton className="h-40 w-full" />
        </div>
      ) : !pendingGames || pendingGames.length === 0 ? (
        <Card className="flex items-center justify-center p-12">
          <p className="text-center text-muted-foreground">
            No games are currently pending setup.
          </p>
        </Card>
      ) : (
        <div className="grid gap-6">
          {pendingGames.map((game) => (
            <GameCard key={game.game_id} game={game} type="pending" />
          ))}
        </div>
      )}
    </section>
  );
}
