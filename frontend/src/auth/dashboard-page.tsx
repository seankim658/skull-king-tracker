import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";

export function DashboardPage() {
  const { user, performLogout } = useAuth(); //

  return (
    <div className="flex flex-col items-center justify-center">
      <div className="w-full max-w-2xl rounded-lg border bg-card text-card-foreground shadow-sm p-6">
        <h1 className="text-3xl font-semibold mb-6 text-center">
          Welcome to Skull King,{" "}
          {user?.display_name || user?.username || "Player"}!
        </h1>
        <p className="text-center mb-4">You are successfully logged in.</p>

        {user && (
          <div className="mb-6 bg-muted p-4 rounded-md text-sm">
            <h2 className="text-lg font-medium mb-2">User Data:</h2>
            <pre className="overflow-x-auto whitespace-pre-wrap break-all">
              {JSON.stringify(user, null, 2)}
            </pre>
          </div>
        )}

        <Button
          onClick={async () => {
            await performLogout();
            // The AuthProvider should handle navigation after logout,
            // but you can add an explicit navigate('/') here if needed.
          }}
          variant="destructive"
          className="w-full"
        >
          Logout
        </Button>
      </div>
    </div>
  );
}
