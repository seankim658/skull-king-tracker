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
import { Badge } from "../ui/badge";
import { useSubmit } from "@/hooks/use-submit";
import { gameAPI } from "@/lib/api/service/game";
import { useQueryClient } from "@tanstack/react-query";
import { useConfirm } from "@/hooks/use-confirm";

interface ScorecardInputDrawerProps {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  currentRound: RoundScorecard;
  players: GamePlayerResponse[];
  editMode: "bids" | "tricks" | null;
  roundToEdit?: RoundScorecard | null;
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
  editMode,
  roundToEdit,
}: ScorecardInputDrawerProps) {
  const { gameId } = useParams<{ gameId: string }>();
  const queryClient = useQueryClient();
  const confirm = useConfirm();
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [bids, setBids] = useState<Record<string, number>>({});
  const [tricks, setTricks] = useState<Record<string, number>>({});
  const [bonus, setBonus] = useState<Record<string, number>>({});

  const onMutationSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ["scorecard", gameId] });
    queryClient.invalidateQueries({ queryKey: ["activeGames"] });
    queryClient.invalidateQueries({ queryKey: ["gameHistory"] });
    queryClient.invalidateQueries({ queryKey: ["globalLeaderboard"] });
    players.forEach((player) => {
      if (player.user_id) {
        queryClient.invalidateQueries({
          queryKey: ["userProfile", player.user_id],
        });
        queryClient.invalidateQueries({
          queryKey: ["userAwardsStats", player.user_id],
        });
      }
    });
    onOpenChange(false);
  };

  const { submit: submitBids, isLoading: isSubmittingBids } = useSubmit(
    gameAPI.submitBids,
    { actionVerb: "Submitting bids", onSuccess: onMutationSuccess },
  );
  const { submit: updateBids, isLoading: isUpdatingBids } = useSubmit(
    gameAPI.updateBids,
    { actionVerb: "Updating bids", onSuccess: onMutationSuccess },
  );
  const { submit: submitTricks, isLoading: isSubmittingTricks } = useSubmit(
    gameAPI.submitTricks,
    { actionVerb: "Submitting scores", onSuccess: onMutationSuccess },
  );
  const { submit: updateTricks, isLoading: isUpdatingTricks } = useSubmit(
    gameAPI.updateTricks,
    { actionVerb: "Updating scores", onSuccess: onMutationSuccess },
  );

  const isLoading =
    isSubmittingBids ||
    isUpdatingBids ||
    isSubmittingTricks ||
    isUpdatingTricks;

  const roundForDisplay = editMode === "tricks" ? roundToEdit : currentRound;
  const phase: InputPhase =
    editMode === "bids"
      ? "bidding"
      : editMode === "tricks"
        ? "tricks"
        : currentRound.status === "bidding"
          ? "bidding"
          : "tricks";
  const activePlayer = players[currentPlayerIndex];

  useEffect(() => {
    if (isOpen) {
      setCurrentPlayerIndex(0);
      // Pre-populate state if in edit mode
      if (editMode === "bids" && currentRound) {
        const initialBids = currentRound.player_scores.reduce(
          (acc, score) => {
            acc[score.game_player_id] = score.bid_amount ?? 0;
            return acc;
          },
          {} as Record<string, number>,
        );
        setBids(initialBids);
      } else if (editMode === "tricks" && roundToEdit) {
        const initialTricks = roundToEdit.player_scores.reduce(
          (acc, score) => {
            acc[score.game_player_id] = score.tricks_taken ?? 0;
            return acc;
          },
          {} as Record<string, number>,
        );
        const initialBonus = roundToEdit.player_scores.reduce(
          (acc, score) => {
            acc[score.game_player_id] = score.bonus_points ?? 0;
            return acc;
          },
          {} as Record<string, number>,
        );
        setTricks(initialTricks);
        setBonus(initialBonus);
      } else {
        // Reset for new entry
        setBids({});
        setTricks({});
        setBonus({});
      }
    }
  }, [isOpen, editMode, currentRound, roundToEdit]);

  const submitData = () => {
    if (phase === "bidding") {
      const payload: SubmitBidsPayload = {
        bids: players.map((p) => ({
          game_player_id: p.game_player_id,
          bid_amount: bids[p.game_player_id] ?? 0,
        })),
      };
      if (editMode === "bids") {
        updateBids(gameId!, currentRound.round_number, payload);
      } else {
        submitBids(gameId!, currentRound.round_number, payload);
      }
    } else {
      const payload: SubmitTricksPayload = {
        tricks: players.map((p) => ({
          game_player_id: p.game_player_id,
          tricks_taken: tricks[p.game_player_id] ?? 0,
          bonus_points: bonus[p.game_player_id] ?? 0,
        })),
      };
      if (editMode === "tricks" && roundToEdit) {
        updateTricks(gameId!, roundToEdit.round_number, payload);
      } else {
        submitTricks(gameId!, currentRound.round_number, payload);
      }
    }
  };

  const handleFinalRoundConfirmation = async () => {
    const summaryContent = (
      <div>
        <p className="mb-4">
          You are submitting scores for the final round. This action cannot be
          undone. Please confirm the scores are correct.
        </p>
        <div className="space-y-2 rounded-md border p-2">
          {players.map((player) => (
            <div
              key={player.game_player_id}
              className="flex justify-between text-sm"
            >
              <span className="font-medium">{player.display_name}:</span>
              <span>
                {tricks[player.game_player_id] ?? 0} Tricks,{" "}
                {bonus[player.game_player_id] ?? 0} Bonus
              </span>
            </div>
          ))}
        </div>
      </div>
    );

    const isConfirmed = await confirm({
      title: "Confirm Final Round Scores?",
      description: summaryContent,
      confirmText: "Submit Final Scores",
      cancelText: "Go Back",
    });

    if (isConfirmed) {
      submitData();
    }
  };

  const handleNext = () => {
    if (currentPlayerIndex < players.length - 1) {
      setCurrentPlayerIndex(currentPlayerIndex + 1);
      return;
    }

    const isFinalRoundSubmission =
      roundForDisplay?.round_number === 10 && phase === "tricks" && !editMode;

    if (isFinalRoundSubmission) {
      handleFinalRoundConfirmation();
    } else {
      submitData();
    }
  };

  const handlePrevious = () => {
    if (currentPlayerIndex > 0) {
      setCurrentPlayerIndex(currentPlayerIndex - 1);
    }
  };

  if (!activePlayer || !roundForDisplay) return null;

  const isLastPlayer = currentPlayerIndex === players.length - 1;
  const currentBid = bids[activePlayer.game_player_id] ?? 0;
  const currentTricks = tricks[activePlayer.game_player_id] ?? 0;
  const currentBonus = bonus[activePlayer.game_player_id] ?? 0;

  const roundForBids = editMode === "tricks" ? roundToEdit : currentRound;
  const bidForTricksPhase =
    phase === "tricks" && roundForBids
      ? (roundForBids.player_scores.find(
          (s) => s.game_player_id === activePlayer.game_player_id,
        )?.bid_amount ?? 0)
      : 0;

  const phaseTitle = editMode
    ? `Edit ${phase === "bidding" ? "Bids" : "Tricks"}`
    : `Enter ${phase === "bidding" ? "Bids" : "Tricks & Bonuses"}`;
  const phaseDescription = `Input for ${activePlayer.display_name}`;
  const phaseSubmit = editMode ? "Save Changes" : "Submit";

  return (
    <Drawer open={isOpen} onOpenChange={onOpenChange}>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader className="text-center">
            <div className="flex items-center justify-center gap-2">
              {editMode && <Badge>Editing</Badge>}
              <DrawerTitle>
                Round {roundForDisplay.round_number} - {phaseTitle}
              </DrawerTitle>
            </div>
            <DrawerDescription>{phaseDescription}</DrawerDescription>
          </DrawerHeader>

          <div className="p-2 flex flex-col items-center gap-4">
            <UserAvatar
              key={`${activePlayer.game_player_id}-${activePlayer.avatar_url || "no-avatar"}`}
              userId={
                activePlayer.user_id
                  ? activePlayer.user_id
                  : activePlayer.guest_player_id
              }
              displayName={activePlayer.display_name}
              avatarUrl={activePlayer.user_id ? activePlayer.avatar_url : null}
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

          <DrawerFooter className="pb-10">
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
