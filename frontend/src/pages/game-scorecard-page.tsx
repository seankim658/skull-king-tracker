import { useParams } from "react-router-dom";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";

export function GameScorecardPage() {
  const { gameId } = useParams<{ gameId: string }>();

  return (
    <div className="container mx-auto p-4 md:p-6">
      <Card>
        <CardHeader>
          <CardTitle>Game Scorecard</CardTitle>
          <CardDescription>Game ID: {gameId}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            (Scorecard Feature Coming Soon)
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
