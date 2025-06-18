import type { ReactNode } from "react";
import type { AuthContextType } from "@/contexts/auth-context";
import type { User } from "@/lib/api/types";
import { useState, useEffect, useCallback, useMemo } from "react";
import { AuthContext } from "@/contexts/auth-context";
import { authAPI } from "@/lib/api/service/auth";
import { useNavigate } from "react-router-dom";
import { useSubmit } from "@/hooks/use-submit";

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
    checkAuthStatus();
  }, [checkAuthStatus]);

  const handleLogoutCompletion = useCallback(() => {
    setUser(null);
    setIsAuthenticated(false);
    navigate("/login", { replace: true });
  }, [navigate]);

  const { submit: logoutAction, isLoading: isLoggingOut } = useSubmit(
    authAPI.logout,
    {
      actionVerb: "Logging out",
      successMessage: "Successfully logged out",
      onSuccess: handleLogoutCompletion,
      onError: handleLogoutCompletion,
    },
  );

  const performLogout = useCallback(async () => {
    await logoutAction();
  }, [logoutAction]);

  const contextValue = useMemo<AuthContextType>(
    () => ({
      user,
      isAuthenticated,
      isLoadingAuth: isLoadingAuth || isLoggingOut,
      checkAuthStatus,
      performLogout,
    }),
    [
      user,
      isAuthenticated,
      isLoadingAuth,
      isLoggingOut,
      checkAuthStatus,
      performLogout,
    ],
  );

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}
