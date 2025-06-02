import { cn } from "@/lib/utils";
import { Link } from "react-router-dom";
import type { LucideIcon } from "lucide-react";
import { CirclePlay, NotebookPen } from "lucide-react";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "../ui/sidebar";

interface NavMainProps {
  items: {
    title: string;
    url: string;
    icon?: LucideIcon;
  }[];
  uploadButton?: boolean;
}

export function NavMain({ items, uploadButton = false }: NavMainProps) {
  return (
    <SidebarGroup>
      <SidebarGroupContent className={cn("flex flex-col gap-2")}>
        {uploadButton && (
          <SidebarMenu>
            <SidebarMenuItem className={cn("flex items-center gap-2")}>
              <SidebarMenuButton
                tooltip="Start Session"
                className={cn(
                  "text-white bg-gradient-to-br from-green-400 to-blue-600 hover:bg-gradient-to-bl",
                )}
                asChild
              >
                <Link to="/start-session">
                  <NotebookPen />
                  <span>Start Session</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem className={cn("flex items-center gap-2")}>
              <SidebarMenuButton
                tooltip="Start Game"
                className={cn(
                  "min-w-8 bg-primary text-primary-foreground duration-200 ease-linear hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground",
                )}
                asChild
              >
                <Link to="/start-game">
                  <CirclePlay />
                  <span>Start Game</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        )}

        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton tooltip={item.title} asChild>
                <Link to={item.url}>
                  {item.icon && <item.icon />}
                  <span>{item.title}</span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
