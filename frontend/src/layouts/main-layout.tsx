import { useMemo } from "react";
import { Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SiteHeader } from "@/components/ui/site-header";
import { getFullAvatarURL } from "@/lib/utils";
import { SiteAlertBanner } from "@/components/ui/site-alert-banner";

export function MainLayout() {
  const { user } = useAuth();

  const sidebarUserData = useMemo(
    () =>
      user
        ? {
            user_id: user.user_id,
            name: user.display_name || user.username,
            email: user.email ? user.email : "",
            avatar: getFullAvatarURL(user.avatar_url),
            updatedAt: user.updated_at,
          }
        : {
            user_id: "",
            name: "Loading...",
            email: "...",
            avatar: "",
            updatedAt: null,
          },
    [user],
  );

  return (
    <SidebarProvider>
      <AppSidebar user={sidebarUserData} variant="inset" />
      <SidebarInset className="flex flex-col min-w-0">
        <div className="flex h-screen flex-col">
          <SiteHeader />
          <main className="flex-1 overflow-auto">
            <SiteAlertBanner />
            <Outlet />
          </main>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
