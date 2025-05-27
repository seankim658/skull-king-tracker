import type { ReactNode } from "react";
import type { AuthContextType } from "@/contexts/auth-context";
import type { User } from "@/lib/api/types";
import { useState, useEffect, useCallback, useMemo } from "react";
import { AuthContext } from "@/contexts/auth-context";
import { authAPI } from "@/lib/api/service/auth";
import { errorExtract } from "@/lib/utils";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoadingAuth, setIsLoadingAuth] = useState<boolean>(true);
  const navigate = useNavigate();

  /**
   * Checks the current authentication status by fetching user data.
   */
  const checkAuthStatus = useCallback(async () => {
    setIsLoadingAuth(true);
    try {
      const response = await authAPI.getCurrentUser();
      if (response.success && response.data?.user) {
        setUser(response.data.user);
        setIsAuthenticated(true);
        // TODO : temp logging
        console.log(
          "Auth status check: User authenticated",
          response.data.user,
        );
      } else {
        setUser(null);
        setIsAuthenticated(false);
        console.log(
          "Auth status check: User not authenticated or failed to fetch user",
        );
      }
    } catch (e) {
      console.warn("Auth status check failed or not active session:", e);
      setUser(null);
      setIsAuthenticated(false);
    } finally {
      setIsLoadingAuth(false);
    }
  }, []);

  useEffect(() => {
    // TODO : temp logging
    console.log("AuthProvider mounted, checking auth status...");
    checkAuthStatus();
  }, [checkAuthStatus]);

  const performLogout = useCallback(async () => {
    setIsLoadingAuth(true);
    try {
      await authAPI.logout();
      toast.success("Successfully logged out");
    } catch (e) {
      console.error("Logout failed:", e);
      toast.error(errorExtract(e, "Logout failed"));
    } finally {
      setUser(null);
      setIsAuthenticated(false);
      setIsLoadingAuth(false);
      navigate("/login", { replace: true });
    }
  }, [navigate]);

  const contextValue = useMemo<AuthContextType>(
    () => ({
      user,
      isAuthenticated,
      isLoadingAuth,
      checkAuthStatus,
      performLogout,
    }),
    [user, isAuthenticated, isLoadingAuth, checkAuthStatus, performLogout],
  );

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}
