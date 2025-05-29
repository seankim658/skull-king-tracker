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
import { errorExtract } from "@/lib/utils";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

export function StartSessionPage() {
  const [sessionName, setSessionName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmedSessionName = sessionName.trim();
    if (!trimmedSessionName) {
      toast.error("Session name cannot be empty");
      return;
    }
    setIsLoading(true);
    const toastId = toast.loading("Starting new session...");

    try {
      const response = await gameAPI.createGame({
        session_name: trimmedSessionName,
      });
      if (response.success && response.data?.game_id) {
        toast.success(
          `Session "${sessionName}" created and first game started`,
          { id: toastId },
        );
        // TODO : Navigate to add players page
      } else {
        toast.error(response.message || "Failed to start session", {
          id: toastId,
        });
      }
    } catch (e) {
      const errMsg = errorExtract(e, "Could not start session and game");
      toast.error(errMsg, { id: toastId });
      console.error(errMsg);
    } finally {
      setIsLoading(false);
    }
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
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? "Starting Session..." : "Create Session"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
