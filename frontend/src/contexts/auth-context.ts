import { createContext } from "react";
import type { User } from "@/lib/api/types";

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoadingAuth: boolean;
  checkAuthStatus: () => Promise<void>;
  performLogout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(
  undefined,
);
