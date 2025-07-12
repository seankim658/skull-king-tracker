import { useParams } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { userAPI } from "@/lib/api/service/user";
import { ProfileHeader } from "@/components/profile/profile-header";
import { ProfileStatsSummary } from "@/components/profile/profile-stats-summary";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Terminal } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { FriendsListModal } from "@/components/profile/friends-list-modal";
import { useState } from "react";
import { ProfileAwardsChart } from "@/components/profile/profile-awards-chart";
import { ReportUserModal } from "@/components/profile/report-user-modal";

export function ProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const { user: authenticatedUser, isLoadingAuth } = useAuth();

  const [isFriendsModalOpen, setIsFriendsModalOpen] = useState(false);
  const [isReportModalOpen, setIsReportModalOpen] = useState(false);

  const {
    data: profileData,
    isLoading,
    isError,
    error,
    refetch: fetchProfile,
  } = useQuery({
    queryKey: ["userProfile", userId],
    queryFn: async () => {
      if (!userId) {
        throw new Error("User ID is required");
      }
      const response = await userAPI.getUserProfile(userId);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Failed to fetch user profile");
      }
      return response.data;
    },
    enabled: !!userId,
  });

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

  if (isError) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <Alert variant="destructive">
          <Terminal className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
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
        onActionSuccess={fetchProfile}
        onOpenFriendsList={() => setIsFriendsModalOpen(true)}
        onReportUser={() => setIsReportModalOpen(true)}
      />
      <ProfileStatsSummary
        stats={profileData.stats}
        username={
          profileData.profile.display_name || profileData.profile.username
        }
      />

      {!!profileData.stats && (
        <ProfileAwardsChart userId={profileData.profile.user_id} />
      )}

      <FriendsListModal
        isOpen={isFriendsModalOpen}
        onClose={() => setIsFriendsModalOpen(false)}
        profile={profileData.profile}
      />

      <ReportUserModal
        isOpen={isReportModalOpen}
        onClose={() => setIsReportModalOpen(false)}
        reportedUser={{
          userId: profileData.profile.user_id,
          displayName:
            profileData.profile.display_name || profileData.profile.username,
        }}
      />
    </div>
  );
}
