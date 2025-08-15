import { createContext, type ReactNode } from "react";

export interface ConfirmOptions {
  title: string;
  description: ReactNode;
  confirmText?: string;
  cancelText?: string;
}

export type ConfirmContextType = (options: ConfirmOptions) => Promise<boolean>;

export const ConfirmContext = createContext<ConfirmContextType | undefined>(
  undefined,
);
