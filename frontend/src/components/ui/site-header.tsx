import { Separator } from "./separator";
import { SidebarTrigger } from "./sidebar";
import { ModeToggle } from "../theme/mode-toggle";
import { NotificationBell } from "./notification-bell";
import { Heart } from "lucide-react";
import { Button } from "./button";

export function SiteHeader() {
  return (
    <header className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <h1 className="text-base font-medium">Skull King Tracker</h1>
        <div className="flex flex-1 items-center justify-end space-x-2">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="cursor-pointer"
          >
            <a
              href="https://ko-fi.com/skim658"
              target="_blank"
              rel="noopener noreferrer"
            >
              <Heart className="h-4 w-4 text-red-500" />
              Support
            </a>
          </Button>
          <NotificationBell />
          <ModeToggle />
        </div>
      </div>
    </header>
  );
}
