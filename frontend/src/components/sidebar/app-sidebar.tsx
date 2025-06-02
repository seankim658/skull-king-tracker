import { useMemo } from "react";
import {
  Home,
  ListIcon,
  Settings,
  HelpCircle,
  ChartLine,
  Gamepad,
  NotebookTabs,
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
} from "../ui/sidebar";
import { Link } from "react-router-dom";

const secondaryItems = [
  {
    title: "Settings",
    url: "/settings",
    icon: Settings,
  },
  {
    title: "Report Bug",
    url: "/bug",
    icon: HelpCircle,
  },
];

interface NavUserProps {
  user_id: string;
  name: string;
  email: string;
  avatar: string;
}

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  user: NavUserProps;
}

export function AppSidebar({ user, ...props }: AppSidebarProps) {
  const navItems = useMemo(
    () => [
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
    ],
    [user.user_id],
  );

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <Link to="/">
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
        <NavUser user={user} />
      </SidebarFooter>
    </Sidebar>
  );
}
