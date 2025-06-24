import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { gameAPI } from "@/lib/api/service/game";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useSubmit } from "@/hooks/use-submit";
import type { GameResponse } from "@/lib/api/types";

export function StartSessionPage() {
  const [sessionName, setSessionName] = useState("");
  const navigate = useNavigate();

  const { submit: createSessionAndGame, isLoading } = useSubmit(
    gameAPI.createGame,
    {
      actionVerb: "Creating session",
      onSuccess: (data: GameResponse | undefined) => {
        toast.success(`Session "${sessionName.trim()} created`);
        if (data?.game_id) {
          navigate(`/game/${data.game_id}/add-players`);
        }
      },
    },
  );

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmedSessionName = sessionName.trim();
    if (!trimmedSessionName) {
      toast.error("Session name cannot be empty");
      return;
    }
    createSessionAndGame({ session_name: trimmedSessionName });
  };

  return (
    <div className="container mx-auto flex min-h-[calc(100vh-var(--header-height))] items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Start a New Game Session</CardTitle>
          <CardDescription>
            Give the session a name. The first game of this session will be
            automatically started in this session.
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="sessionName">Session Name</Label>
              <Input
                id="sessionName"
                value={sessionName}
                onChange={(e) => setSessionName(e.target.value)}
                placeholder="E.g., Friday Night Skull King"
                disabled={isLoading}
                required
              />
            </div>
          </CardContent>
          <CardFooter className="pt-4">
            <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
              {isLoading ? "Starting Session..." : "Create Session"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
