import { useState, useEffect } from "react";
import type { FormEvent } from "react";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { useConfirm } from "@/hooks/use-confirm";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ColorThemeSelector } from "@/components/theme/color-theme-selector";
import { ModeSelector } from "@/components/theme/mode-selector";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { UserAvatar } from "@/components/ui/user-avatar";
import { userAPI } from "@/lib/api/service/user";
import type { UpdateUserProfilePayload, User } from "@/lib/api/types";
import { PageHeader } from "@/components/ui/page-header";
import { toast } from "sonner";
import { API_BASE_URL } from "@/lib/api/client";
import { AVAILABLE_OAUTH_PROVIDERS } from "@/lib/providers";
import { useSubmit } from "@/hooks/use-submit";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { InfoTooltip } from "@/components/ui/info-tooltip";

type StatsPrivacy = User["stats_privacy"];
const STATS_PRIVACY_OPTIONS: { value: StatsPrivacy; label: string }[] = [
  { value: "public", label: "Public" },
  { value: "friends_only", label: "Friends Only" },
  { value: "private", label: "Private" },
];

export function SettingsPage() {
  const queryClient = useQueryClient();
  const confirm = useConfirm();
  const { user, checkAuthStatus, isLoadingAuth } = useAuth();

  const [displayName, setDisplayName] = useState("");
  const [avatarUrl, setAvatarUrl] = useState("");
  const [statsPrivacy, setStatsPrivacy] = useState<StatsPrivacy>("public");

  const { data: linkedAccounts, isLoading: isLoadingLinkedAccounts } = useQuery(
    {
      queryKey: ["linkedAccounts"],
      queryFn: async () => {
        const response = await userAPI.getLinkedAccounts();
        if (!response.success || !response.data) {
          throw new Error(
            response.message || "Could not fetch linked accounts",
          );
        }
        return response.data;
      },
      enabled: !!user,
    },
  );

  const { submit: updateProfile, isLoading: isSavingProfile } = useSubmit(
    userAPI.updateUserProfile,
    {
      actionVerb: "Saving profile",
      successMessage: "Profile updated successfully",
      onSuccess: () => {
        if (user) {
          queryClient.invalidateQueries({
            queryKey: ["userProfile", user.user_id],
          });
        }
        checkAuthStatus();
      },
    },
  );

  const { submit: disconnectProvider, isLoading: isDisconnecting } = useSubmit(
    userAPI.unlinkAccount,
    {
      actionVerb: "Disconnecting account",
      successMessage: "Account unlinked successfully",
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["linkedAccounts"] });
      },
    },
  );

  useEffect(() => {
    if (user) {
      setDisplayName(user.display_name || user.username || "");
      setAvatarUrl(user.avatar_url || "");
      setStatsPrivacy(user.stats_privacy);
    }
  }, [user]);

  const handleProfileSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!user) return;

    const payload: UpdateUserProfilePayload = {};
    if (displayName.trim() !== (user.display_name || ""))
      payload.display_name = displayName.trim();
    if (avatarUrl !== (user.avatar_url || "")) payload.avatar_url = avatarUrl;
    if (statsPrivacy !== user.stats_privacy)
      payload.stats_privacy = statsPrivacy;

    if (Object.keys(payload).length === 0) {
      toast.info("No changes to save");
      return;
    }
    updateProfile(payload);
  };

  const handleDisconnectProvider = async (
    providerId: string,
    providerName: string,
  ) => {
    const isConfirmed = await confirm({
      title: `Disconnect ${providerName}?`,
      description: `Are you sure you want to disconnect your ${providerName} account?`,
      confirmText: "Disconnect",
    });
    if (isConfirmed) {
      disconnectProvider(providerId);
    }
  };

  const handleConnectProvider = (providerId: string) => {
    window.location.href = `${API_BASE_URL}/auth/initiate-link/${providerId}`;
  };

  if (isLoadingAuth || !user) {
    return (
      <div className="container mx-auto p-4 md:p-6">
        <h1 className="text-2xl font-bold mb-6">Settings</h1>
        <p>Loading user settings...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-3xl space-y-8 p-4 py-8 md:p-6 lg:py-12">
      <PageHeader
        title="Settings"
        description="Manage your account and theme preferences"
      />

      <Separator />

      {/* Theme Settings */}
      <section className="space-y-6">
        <h2 className="text-xl font-semibold">Appearance</h2>
        <Card>
          <CardHeader>
            <CardTitle>Theme</CardTitle>
            <CardDescription>
              Customize the look and feel of the application
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <Label htmlFor="mode-selector">UI Mode</Label>
              <ModeSelector />
            </div>
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <Label htmlFor="color-theme-selector">Color Theme</Label>
              <ColorThemeSelector />
            </div>
          </CardContent>
        </Card>
      </section>

      <Separator />

      {/* Profile Settings */}
      <section className="space-y-6">
        <h2 className="text-xl font-semibold">Profile</h2>
        <form onSubmit={handleProfileSubmit}>
          <Card>
            <CardHeader>
              <CardTitle>Account Information</CardTitle>
              <CardDescription>
                Update your public profile details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="displayName">Display Name</Label>
                <Input
                  id="displayName"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Enter your display name"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="avatarUrl">Avatar URL</Label>
                <div className="flex items-center gap-4">
                  <UserAvatar
                    displayName={displayName || user.username}
                    avatarUrl={user.avatar_url}
                    className="h-16 w-16"
                  />
                  <Input
                    id="avatarUrl"
                    type="url"
                    value={avatarUrl}
                    onChange={(e) => setAvatarUrl(e.target.value)}
                    placeholder="https://example.com/avatar.png"
                    className="flex-1"
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  Enter a URL for your avatar image. This image will be
                  displayed in a circle.
                </p>
              </div>

              <div className="space-y-2">
                <div className="flex items-start gap-1">
                  <Label htmlFor="stats-privacy">Stats Privacy</Label>
                  <InfoTooltip
                    content={
                      <>
                        <p>
                          <strong>Public:</strong> Anyone can see your stats.
                        </p>
                        <p>
                          <strong>Friends Only:</strong> Only your accepted
                          friends can see your stats.
                        </p>
                        <p>
                          <strong>Private:</strong> Only you can see your stats.
                        </p>
                      </>
                    }
                  />
                </div>
                <Select
                  value={statsPrivacy}
                  onValueChange={(value: StatsPrivacy) =>
                    setStatsPrivacy(value)
                  }
                >
                  <SelectTrigger id="stats-privacy" className="w-full">
                    <SelectValue placeholder="Select your stat privacy level" />
                  </SelectTrigger>
                  <SelectContent align="end">
                    <SelectGroup>
                      {STATS_PRIVACY_OPTIONS.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Control who can see your game statistics.
                </p>
              </div>
            </CardContent>
            <CardFooter>
              <Button type="submit" disabled={isSavingProfile}>
                {isSavingProfile ? "Saving..." : "Save Profile Changes"}
              </Button>
            </CardFooter>
          </Card>
        </form>
      </section>

      <Separator />

      <section className="space-y-6">
        <h2 id="linked-accounts-heading" className="text-xl font-semibold">
          Linked Accounts
        </h2>
        <Card>
          <CardHeader>
            <CardTitle>Manage Authentication Methods</CardTitle>
            <CardDescription>
              Connect or disconnect your OAuth accounts. You must have at least
              one authentication method.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {isLoadingLinkedAccounts && <p>Loading linked accounts...</p>}
            {!isLoadingLinkedAccounts &&
              linkedAccounts &&
              AVAILABLE_OAUTH_PROVIDERS.map((availableProvider) => {
                const linkedAccount = linkedAccounts.find(
                  (acc) => acc.provider_name === availableProvider.id,
                );
                const isOnlyAccount =
                  linkedAccounts.length === 1 && !!linkedAccount;

                return (
                  <div
                    key={availableProvider.id}
                    className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 border rounded-lg gap-3 sm:gap-2"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8">{availableProvider.icon}</div>
                      <div>
                        <span className="font-medium">
                          {availableProvider.name}
                        </span>
                        {linkedAccount && (
                          <p className="text-xs text-muted-foreground">
                            Connected:{" "}
                            {linkedAccount.provider_display_name ||
                              linkedAccount.proivder_email ||
                              "Yes"}
                          </p>
                        )}
                      </div>
                    </div>
                    {linkedAccount ? (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          handleDisconnectProvider(
                            availableProvider.id,
                            availableProvider.name,
                          )
                        }
                        disabled={isOnlyAccount || isDisconnecting}
                        className="w-full sm:w-auto"
                      >
                        Disconnect
                      </Button>
                    ) : (
                      <Button
                        variant="default"
                        size="sm"
                        onClick={() =>
                          handleConnectProvider(availableProvider.id)
                        }
                        className="w-full sm:w-auto"
                      >
                        Connect
                      </Button>
                    )}
                  </div>
                );
              })}
            {!isLoadingLinkedAccounts && linkedAccounts?.length === 0 && (
              <p className="text-sm text-muted-foreground py-4 text-center">
                No linked accounts found.
              </p>
            )}
          </CardContent>
          {linkedAccounts && linkedAccounts.length === 1 && (
            <CardFooter>
              <p className="text-xs text-muted-foreground">
                You cannot disconnect your only authentication method. Please
                connect another account first if you wish to remove this one.
              </p>
            </CardFooter>
          )}
        </Card>
      </section>
    </div>
  );
}
