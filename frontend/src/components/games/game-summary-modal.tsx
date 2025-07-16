import { useQuery } from "@tanstack/react-query";
import { gameAPI } from "@/lib/api/service/game";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "../ui/dialog";
import { Skeleton } from "../ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { UserAvatar } from "../ui/user-avatar";
import { Terminal, Award } from "lucide-react";
import { Separator } from "../ui/separator";
import type { GameAward } from "@/lib/api/types";
import { StatCard } from "../ui/stat-card";

interface GameSummaryModalProps {
  gameId: string | null;
  isOpen: boolean;
  onClose: () => void;
}

const AwardCard = ({ award }: { award: GameAward }) => (
  <StatCard
    title={award.title}
    value={award.player_name}
    icon={<Award className="h-5 w-5 text-amber-500" />}
    description={award.value}
    tooltip={award.description}
  />
);

export function GameSummaryModal({
  gameId,
  isOpen,
  onClose,
}: GameSummaryModalProps) {
  const {
    data: summaryData,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["gameSummary", gameId],
    queryFn: async () => {
      const response = await gameAPI.getGameSummary(gameId!);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch game summary");
      }
      return response.data;
    },
    enabled: !!gameId && isOpen,
  });

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="space-y-6">
          <Skeleton className="h-24 w-full" />
          <div className="grid grid-cols-2 gap-4">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-20" />
            ))}
          </div>
          <Skeleton className="h-32 w-full" />
        </div>
      );
    }

    if (isError) {
      return (
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error Loading Summary</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      );
    }

    if (!summaryData) return null;

    const topThree = summaryData.final_scores
      .slice()
      .sort((a, b) => b.final_score - a.final_score)
      .slice(0, 3);

    return (
      <div className="space-y-6">
        {/* Podium */}
        <div className="grid grid-cols-3 items-end gap-2 p-4 bg-muted/50 rounded-lg">
          {/* 2nd Place Slot */}
          <div className="text-center">
            {topThree[1] && (
              <>
                <p className="text-2xl">🥈</p>
                <UserAvatar
                  userId={topThree[1].user_id}
                  displayName={topThree[1].display_name}
                  avatarUrl={topThree[1].avatar_url}
                  className="h-16 w-16 mx-auto border-2 border-slate-400"
                />
                <p className="font-semibold mt-1 truncate">
                  {topThree[1].display_name}
                </p>
                <p className="text-sm text-muted-foreground">
                  {topThree[1].final_score} pts
                </p>
              </>
            )}
          </div>
          {/* 1st Place */}
          <div className="text-center">
            {topThree[0] && (
              <>
                <p className="text-4xl">🥇</p>
                <UserAvatar
                  userId={topThree[0].user_id}
                  displayName={topThree[0].display_name}
                  avatarUrl={topThree[0].avatar_url}
                  className="h-20 w-20 mx-auto border-4 border-amber-400"
                />
                <p className="text-lg font-bold mt-1 truncate">
                  {topThree[0].display_name}
                </p>
                <p className="text-muted-foreground">
                  {topThree[0].final_score} pts
                </p>
              </>
            )}
          </div>
          {/* 3rd Place Slot */}
          <div className="text-center">
            {topThree[2] && (
              <>
                <p className="text-2xl">🥉</p>
                <UserAvatar
                  userId={topThree[2].user_id}
                  displayName={topThree[2].display_name}
                  avatarUrl={topThree[2].avatar_url}
                  className="h-16 w-16 mx-auto border-2 border-amber-600"
                />
                <p className="font-semibold mt-1 truncate">
                  {topThree[2].display_name}
                </p>
                <p className="text-sm text-muted-foreground">
                  {topThree[2].final_score} pts
                </p>
              </>
            )}
          </div>
        </div>

        {/* Awards */}
        {summaryData.awards.length > 0 && (
          <div>
            <h3 className="text-lg font-semibold mb-3 text-center">
              Game Awards
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {summaryData.awards.map((award) => (
                <AwardCard key={award.title} award={award} />
              ))}
            </div>
          </div>
        )}

        <Separator />

        {/* Final Scores */}
        <div>
          <h3 className="text-lg font-semibold mb-2">Final Scoreboard</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[50px]">Pos</TableHead>
                <TableHead>Player</TableHead>
                <TableHead className="text-right">Scores</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {summaryData.final_scores.map((player, idx) => (
                <TableRow key={player.game_player_id}>
                  <TableCell className="font-medium">{idx + 1}</TableCell>
                  <TableCell>{player.display_name}</TableCell>
                  <TableCell className="text-right">
                    {player.final_score}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-2xl text-center">Game Over!</DialogTitle>
          <DialogDescription className="text-center">
            Final game summary
          </DialogDescription>
        </DialogHeader>
        <div className="py-2">{renderContent()}</div>
      </DialogContent>
    </Dialog>
  );
}
