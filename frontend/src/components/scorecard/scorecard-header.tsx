import { CardHeader, CardTitle, CardDescription } from "../ui/card";
import { NotebookPen } from "lucide-react";

interface ScorecardHeaderProps {
  sessionName?: string | null;
  gameId: string;
  scorekeeperName: string;
}

export function ScorecardHeader({
  sessionName,
  gameId,
  scorekeeperName,
}: ScorecardHeaderProps) {
  return (
    <CardHeader>
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
        <div>
          <CardTitle className="text-2xl">
            {sessionName || "One-off Game"}
          </CardTitle>
          <CardDescription>Game ID: {gameId}</CardDescription>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground mt-2 sm:mt-0">
          <NotebookPen className="h-4 w-4" />
          <span>Scorekeeper: {scorekeeperName}</span>
        </div>
      </div>
    </CardHeader>
  );
}
