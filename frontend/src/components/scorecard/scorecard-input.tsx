import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import type { ReactNode } from "react";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerFooter,
  DrawerDescription,
} from "../ui/drawer";
import { Button } from "../ui/button";
import { UserAvatar } from "../ui/user-avatar";
import type {
  GamePlayerResponse,
  RoundScorecard,
  SubmitBidsPayload,
  SubmitTricksPayload,
} from "@/lib/api/types";
import { RotateCcw } from "lucide-react";
import { useSubmit } from "@/hooks/use-submit";
import { gameAPI } from "@/lib/api/service/game";
import { useQueryClient } from "@tanstack/react-query";

interface ScorecardInputDrawerProps {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  currentRound: RoundScorecard;
  players: GamePlayerResponse[];
}

type InputPhase = "bidding" | "tricks";

// Helper for quick-tap buttons
const ActionButton = ({
  onClick,
  children,
}: {
  onClick: () => void;
  children: ReactNode;
}) => (
  <Button
    variant="outline"
    size="lg"
    className="h-14 text-lg cursor-pointer"
    onClick={onClick}
  >
    {children}
  </Button>
);

export function ScorecardInputDrawer({
  isOpen,
  onOpenChange,
  currentRound,
  players,
}: ScorecardInputDrawerProps) {
  const { gameId } = useParams<{ gameId: string }>();
  const queryClient = useQueryClient();
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [bids, setBids] = useState<Record<string, number>>({});
  const [tricks, setTricks] = useState<Record<string, number>>({});
  const [bonus, setBonus] = useState<Record<string, number>>({});

  const { submit: submitBids, isLoading: isSubmittingBids } = useSubmit(
    gameAPI.submitBids,
    {
      actionVerb: "Submitting bids",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["scorecard", gameId] });
        onOpenChange(false);
      },
    },
  );

  const { submit: submitTricks, isLoading: isSubmittingTricks } = useSubmit(
    gameAPI.submitTricks,
    {
      actionVerb: "Submitting scores",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["scorecard", gameId] });
        queryClient.invalidateQueries({ queryKey: ["activeGames"] });
        queryClient.invalidateQueries({ queryKey: ["gameHistory"] });
        onOpenChange(false);
      },
    },
  );

  const isLoading = isSubmittingBids || isSubmittingTricks;

  const phase: InputPhase =
    currentRound.status === "bidding" ? "bidding" : "tricks";
  const activePlayer = players[currentPlayerIndex];

  useEffect(() => {
    if (isOpen) {
      setCurrentPlayerIndex(0);
      setBids({});
      setTricks({});
      setBonus({});
    }
  }, [isOpen, currentRound.round_number]);

  const handleNext = () => {
    if (currentPlayerIndex < players.length - 1) {
      setCurrentPlayerIndex(currentPlayerIndex + 1);
      return;
    }

    if (phase === "bidding") {
      const payload: SubmitBidsPayload = {
        bids: players.map((p) => ({
          game_player_id: p.game_player_id,
          bid_amount: bids[p.game_player_id] ?? 0,
        })),
      };
      submitBids(gameId!, currentRound.round_number, payload);
    } else {
      const payload: SubmitTricksPayload = {
        tricks: players.map((p) => ({
          game_player_id: p.game_player_id,
          tricks_taken: tricks[p.game_player_id] ?? 0,
          bonus_points: bonus[p.game_player_id] ?? 0,
        })),
      };
      submitTricks(gameId!, currentRound.round_number, payload);
    }
  };

  const handlePrevious = () => {
    if (currentPlayerIndex > 0) {
      setCurrentPlayerIndex(currentPlayerIndex - 1);
    }
  };

  if (!activePlayer) return null;

  const isLastPlayer = currentPlayerIndex === players.length - 1;
  const currentBid = bids[activePlayer.game_player_id] ?? 0;
  const currentTricks = tricks[activePlayer.game_player_id] ?? 0;
  const currentBonus = bonus[activePlayer.game_player_id] ?? 0;

  const bidForTricksPhase =
    phase === "tricks"
      ? (currentRound.player_scores.find(
          (s) => s.game_player_id === activePlayer.game_player_id,
        )?.bid_amount ?? 0)
      : 0;

  const phaseTitle =
    phase === "bidding" ? "Enter Bids" : "Enter Tricks & Bonuses";
  const phaseDescription =
    phase === "bidding" ? "Input bids for " : "Input results for ";
  const phaseSubmit = phase === "bidding" ? "Submit Bids" : "Submit Scores";

  return (
    <Drawer open={isOpen} onOpenChange={onOpenChange}>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader className="text-center">
            <DrawerTitle>
              Round {currentRound.round_number} - {phaseTitle}
            </DrawerTitle>
            <DrawerDescription>
              {`${phaseDescription}${activePlayer.display_name}`}
            </DrawerDescription>
          </DrawerHeader>

          <div className="p-2 flex flex-col items-center gap-4">
            <UserAvatar
              displayName={activePlayer.display_name}
              avatarUrl={activePlayer.avatar_url}
              className="h-16 w-16"
            />

            {phase === "bidding" ? (
              <div className="w-full space-y-4">
                <div className="text-center p-4 bg-muted rounded-lg">
                  <p className="text-7xl font-bold tracking-tighter">
                    {currentBid}
                  </p>
                  <p className="text-muted-foreground">Bid</p>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  <ActionButton
                    onClick={() =>
                      setBids((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: Math.max(
                          0,
                          currentBid - 2,
                        ),
                      }))
                    }
                  >
                    -2
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBids((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: Math.max(
                          0,
                          currentBid - 1,
                        ),
                      }))
                    }
                  >
                    -1
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBids((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: 0,
                      }))
                    }
                  >
                    <RotateCcw className="h-4 w-4" />
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBids((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBid + 1,
                      }))
                    }
                  >
                    +1
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBids((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBid + 2,
                      }))
                    }
                  >
                    +2
                  </ActionButton>
                </div>
              </div>
            ) : (
              <div className="w-full space-y-4">
                <div className="text-center">
                  <p className="text-sm text-muted-foreground">
                    Player Bid:
                    <span className="font-bold text-lg text-foreground ml-2">
                      {bidForTricksPhase}
                    </span>
                  </p>
                </div>
                <div className="text-center p-2 bg-muted rounded-lg">
                  <p className="text-5xl font-bold tracking-tighter">
                    {currentTricks}
                  </p>
                  <p className="text-xs text-muted-foreground">Tricks Won</p>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  <ActionButton
                    onClick={() =>
                      setTricks((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: Math.max(
                          0,
                          currentTricks - 2,
                        ),
                      }))
                    }
                  >
                    -2
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setTricks((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: Math.max(
                          0,
                          currentTricks - 1,
                        ),
                      }))
                    }
                  >
                    -1
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setTricks((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: 0,
                      }))
                    }
                  >
                    <RotateCcw className="h-4 w-4" />
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setTricks((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentTricks + 1,
                      }))
                    }
                  >
                    +1
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setTricks((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentTricks + 2,
                      }))
                    }
                  >
                    +2
                  </ActionButton>
                </div>

                <div className="text-center p-2 bg-muted rounded-lg">
                  <p className="text-5xl font-bold tracking-tighter">
                    {currentBonus}
                  </p>
                  <p className="text-xs text-muted-foreground">Bonus Points</p>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  <ActionButton
                    onClick={() =>
                      setBonus((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: 0,
                      }))
                    }
                  >
                    <RotateCcw className="h-4 w-4" />
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBonus((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBonus + 10,
                      }))
                    }
                  >
                    +10
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBonus((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBonus + 20,
                      }))
                    }
                  >
                    +20
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBonus((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBonus + 40,
                      }))
                    }
                  >
                    +40
                  </ActionButton>
                  <ActionButton
                    onClick={() =>
                      setBonus((p) => ({
                        ...p,
                        [activePlayer.game_player_id]: currentBonus + 50,
                      }))
                    }
                  >
                    +50
                  </ActionButton>
                </div>
              </div>
            )}
          </div>

          <DrawerFooter>
            <Button
              variant="outline"
              onClick={handlePrevious}
              size="lg"
              className="cursor-pointer"
              disabled={currentPlayerIndex === 0 || isLoading}
            >
              Previous Player
            </Button>
            <Button
              onClick={handleNext}
              size="lg"
              className="cursor-pointer"
              disabled={isLoading}
            >
              {isLastPlayer ? phaseSubmit : "Next Player"}
            </Button>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  );
}
