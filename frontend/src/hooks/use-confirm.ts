import { useContext } from "react";
import {
  ConfirmContext,
  type ConfirmContextType,
} from "@/contexts/confirm-context";

export function useConfirm(): ConfirmContextType {
  const context = useContext(ConfirmContext);
  if (context === undefined) {
    throw new Error("useConfirm must be used within a ConfirmProvider");
  }
  return context;
}
