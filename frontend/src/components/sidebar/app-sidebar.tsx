import { useMemo } from "react";
import {
  Home,
  ListIcon,
  Settings,
  HelpCircle,
  Shield,
  ChartLine,
  Gamepad,
  NotebookTabs,
  Lightbulb,
} from "lucide-react";
import { NavUser } from "./nav-user";
import { NavMain } from "./nav-main";
import { NavSecondary } from "./nav-secondary";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "../ui/sidebar";
import { Link } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";

const secondaryItems = [
  {
    title: "Settings",
    url: "/settings",
    icon: Settings,
  },
  {
    title: "Report Bug",
    url: "https://github.com/seankim658/skull-king-tracker/issues/new/choose",
    icon: HelpCircle,
    external: true,
  },
  {
    title: "Feature Request",
    url: "https://github.com/seankim658/skull-king-tracker/issues/new/choose",
    icon: Lightbulb,
    external: true,
  },
];

interface NavUserProps {
  user_id: string;
  name: string;
  email: string;
  avatar: string | undefined;
  updatedAt: string | null;
}

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  user: NavUserProps;
}

export function AppSidebar({ user, ...props }: AppSidebarProps) {
  const { user: authenticatedUser } = useAuth();
  const { setOpen } = useSidebar();

  const handleLinkClick = () => {
    setOpen(false);
  };

  const navItems = useMemo(() => {
    const items = [
      {
        title: "Home",
        url: "/",
        icon: Home,
      },
      {
        title: "Sessions",
        url: "/sessions",
        icon: ListIcon,
      },
      {
        title: "Games",
        url: "/games",
        icon: Gamepad,
      },
      {
        title: "My Profile",
        url: user.user_id ? `/users/${user.user_id}` : "/login",
        icon: NotebookTabs,
      },
      {
        title: "Explore",
        url: "/explore",
        icon: ChartLine,
      },
    ];

    if (authenticatedUser?.role === "superuser") {
      items.push({
        title: "Admin Panel",
        url: "/admin/dashboard",
        icon: Shield,
      });
    }

    return items;
  }, [user.user_id, authenticatedUser?.role]);

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <Link to="/" onClick={handleLinkClick}>
                <span className="text-base font-semibold">
                  Skull King Tracker
                </span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navItems} uploadButton />
        <NavSecondary items={secondaryItems} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={{ ...user, avatar: user.avatar || "" }} />
      </SidebarFooter>
    </Sidebar>
  );
}
