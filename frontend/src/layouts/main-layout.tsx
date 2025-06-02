import { useMemo } from "react";
import { Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SiteHeader } from "@/components/ui/site-header";
import { getFullAvatarURL } from "@/lib/utils";

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
          }
        : { user_id: "", name: "Loading...", email: "...", avatar: "" },
    [user],
  );

  return (
    <SidebarProvider>
      <AppSidebar user={sidebarUserData} variant="inset" />
      <SidebarInset>
        <SiteHeader />
        {/* Main content area where child routes will render */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
