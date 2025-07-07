import { Tooltip, TooltipTrigger, TooltipContent } from "../ui/tooltip";
import { Info } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface InfoTooltipProps {
  content: ReactNode;
  wrapperClassName?: string;
  iconClassName?: string;
  contentClassName?: string;
  position?: "absolute" | "inline";
}

export function InfoTooltip({
  content,
  wrapperClassName,
  iconClassName,
  contentClassName,
  position = "absolute",
}: InfoTooltipProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <div
          className={cn(
            position === "absolute" &&
              "absolute top-0 right-0 translate-x-1/2 -translate-y-1/4",
            position === "inline" && "inline-flex items-center",
            wrapperClassName,
          )}
        >
          <Info
            className={cn("h-3 w-3 text-muted-foreground", iconClassName)}
          />
        </div>
      </TooltipTrigger>
      <TooltipContent className={cn("max-w-xs", contentClassName)}>
        <div className="text-wrap">{content}</div>
      </TooltipContent>
    </Tooltip>
  );
}
