import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { API_BASE_URL } from "@/lib/api/client";
import { AVAILABLE_OAUTH_PROVIDERS } from "@/lib/providers";
import type { ReactNode } from "react";

interface LoginFormProps {
  className?: string;
  headerContent?: ReactNode;
}

export function LoginForm({
  className,
  headerContent,
  ...props
}: LoginFormProps) {
  /**
   * Handles the login click by redirecting to the backend's OAuth endpoint.
   */
  const handleLogin = (providerId: string) => {
    window.location.href = `${API_BASE_URL}/auth/${providerId}/login`;
  };

  return (
    <div
      className={cn(
        "flex min-h-screen flex-col items-center justify-center bg-background p-4 md:p-6",
        className,
      )}
      {...props}
    >
      {headerContent && (
        <div className="mb-6 text-center md:mb-8">{headerContent}</div>
      )}

      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Welcome back</CardTitle>
          <CardDescription>
            Login to start tracking your Skull King games
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4">
            {AVAILABLE_OAUTH_PROVIDERS.map((provider) => (
              <Button
                key={provider.id}
                variant="default"
                className="w-full cursor-pointer flex items-center justify-center gap-2"
                onClick={() => handleLogin(provider.id)}
              >
                <div className="h-5 w-5">{provider.icon}</div>
                Login with {provider.name}
              </Button>
            ))}
          </div>
          <div className="mt-6 text-center text-xs text-muted-foreground">
            By clicking continue, you agree to our{" "}
            <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>.
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
