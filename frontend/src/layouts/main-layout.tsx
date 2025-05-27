import { Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/sidebar/app-sidebar";
import { SiteHeader } from "@/components/ui/site-header";

export function MainLayout() {
  const { user } = useAuth();

  const sidebarUserData = user
    ? {
        name: user.display_name || user.username,
        email: user.email ? user.email : "",
        avatar: user.avatar_url ? user.avatar_url : "",
      }
    : { name: "Loading...", email: "...", avatar: "" };

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
