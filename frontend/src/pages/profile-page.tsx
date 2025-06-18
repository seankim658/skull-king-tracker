import { useEffect, } from "react";
import { useParams } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { userAPI } from "@/lib/api/service/user";
import { ProfileHeader } from "@/components/profile/profile-header";
import { ProfileStatsSummary } from "@/components/profile/profile-stats-summary";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Terminal } from "lucide-react";
import { useApi } from "@/hooks/use-api";

export function ProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const { user: authenticatedUser, isLoadingAuth } = useAuth();

  const {
    data: profileData,
    isLoading,
    error,
    request: fetchProfile,
  } = useApi(userAPI.getUserProfile);

  useEffect(() => {
    if (userId) {
      fetchProfile(userId);
    }
  }, [userId, fetchProfile]);

  if (isLoading || isLoadingAuth) {
    return (
      <div className="container mx-auto p-4 md:p-6 space-y-8">
        <Skeleton className="h-32 w-full" />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!profileData) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Alert>
          <Terminal className="h-4 w-4" />
          <AlertTitle>Profile Not Found</AlertTitle>
          <AlertDescription>
            The requested user profile could not be found.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const isOwnProfile =
    authenticatedUser?.user_id === profileData.profile.user_id;

  return (
    <div className="container mx-auto p-4 md:p-6 space-y-8">
      <ProfileHeader
        profile={profileData.profile}
        isOwnProfile={isOwnProfile}
        onActionSuccess={() => userId && fetchProfile(userId)}
      />
      <ProfileStatsSummary
        stats={profileData.stats}
        username={
          profileData.profile.display_name || profileData.profile.username
        }
      />
      {/* TODO Future sections like Game History, Detailed Stats can be added here */}
    </div>
  );
}
