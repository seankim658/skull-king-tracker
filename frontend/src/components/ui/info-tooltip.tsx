import { Tooltip, TooltipTrigger, TooltipContent } from "../ui/tooltip";
import { Button } from "./button";
import { Info } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface InfoTooltipProps {
  content: ReactNode;
  buttonClassName?: string;
  iconClassName?: string;
  contentClassName?: string;
}

export function InfoTooltip({
  content,
  buttonClassName,
  iconClassName,
  contentClassName,
}: InfoTooltipProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className={cn("h-4 w-4 rounded-full flex-shrink-0", buttonClassName)}
        >
          <Info
            className={cn("h-3 w-3 text-muted-foreground", iconClassName)}
          />
        </Button>
      </TooltipTrigger>
      <TooltipContent className={cn("max-w-xs", contentClassName)}>
        <div className="text-wrap">{content}</div>
      </TooltipContent>
    </Tooltip>
  );
}
