import { cn } from "@/lib/utils";
import { Link, useNavigate } from "react-router-dom";
import type { LucideIcon } from "lucide-react";
import { CirclePlay, NotebookPen } from "lucide-react";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "../ui/sidebar";
import { useSubmit } from "@/hooks/use-submit";
import { gameAPI } from "@/lib/api/service/game";
import type { GameResponse } from "@/lib/api/types";

interface NavMainProps {
  items: {
    title: string;
    url: string;
    icon?: LucideIcon;
  }[];
  uploadButton?: boolean;
}

export function NavMain({ items, uploadButton = false }: NavMainProps) {
  const { setOpen, isMobile } = useSidebar();
  const navigate = useNavigate();

  const handleLinkClick = () => {
    if (isMobile) {
      setOpen(false);
    }
  };

  const { submit: createOneOffGame, isLoading: isCreatingGame } = useSubmit(
    gameAPI.createGame,
    {
      actionVerb: "Starting game",
      successMessage: "New one-off game started",
      onSuccess: (data: GameResponse | undefined) => {
        if (data?.game_id) {
          navigate(`/game/${data.game_id}/add-players`);
          handleLinkClick();
        }
      },
    },
  );

  const handleStartGameClick = () => {
    createOneOffGame({});
  };

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
                onClick={handleStartGameClick}
                disabled={isCreatingGame}
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
                <Link to={item.url} onClick={handleLinkClick}>
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
