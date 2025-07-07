"use client";

import * as React from "react";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { cn } from "@/lib/utils";
import { useIsMobile } from "@/hooks/use-mobile";

// --- Popover Components (for mobile) ---
//
const Popover = PopoverPrimitive.Root;
const PopoverTrigger = PopoverPrimitive.Trigger;
const PopoverContent = React.forwardRef<
  React.ElementRef<typeof PopoverPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(({ className, align = "center", sideOffset = 4, ...props }, ref) => (
  <PopoverPrimitive.Portal>
    <PopoverPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={cn(
        "z-50 w-fit rounded-lg border bg-primary p-3 text-popover-foreground shadow-md outline-none animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95",
        className,
      )}
      {...props}
    />
  </PopoverPrimitive.Portal>
));
PopoverContent.displayName = PopoverPrimitive.Content.displayName;

// --- Tooltip Components (for desktop) ---
//
const TooltipProvider = TooltipPrimitive.Provider;

const TooltipContext = React.createContext({ isMobile: false });

const useTooltipContext = () => {
  const context = React.useContext(TooltipContext);
  if (!context) {
    throw new Error("Tooltip components must be used within a <Tooltip>");
  }
  return context;
};

function Tooltip({ ...props }) {
  const isMobile = useIsMobile();
  const contextValue = { isMobile };

  return (
    <TooltipContext.Provider value={contextValue}>
      {isMobile ? (
        <Popover {...props} />
      ) : (
        <TooltipProvider>
          <TooltipPrimitive.Root delayDuration={100} {...props} />
        </TooltipProvider>
      )}
    </TooltipContext.Provider>
  );
}

const TooltipTrigger = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Trigger>
>(({ className, ...props }, ref) => {
  const { isMobile } = useTooltipContext();
  const TriggerComponent = isMobile ? PopoverTrigger : TooltipPrimitive.Trigger;
  return <TriggerComponent ref={ref} className={className} {...props} />;
});
TooltipTrigger.displayName = "TooltipTrigger";

// --- Content Component ---
const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => {
  const { isMobile } = useTooltipContext();

  if (isMobile) {
    return (
      <PopoverContent
        ref={ref}
        sideOffset={sideOffset}
        className={className}
        {...props}
      />
    );
  }

  return (
    <TooltipPrimitive.Content
      ref={ref}
      sideOffset={sideOffset}
      className={cn(
        "z-50 overflow-hidden rounded-md bg-primary px-3 py-1.5 text-xs text-primary-foreground animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
        className,
      )}
      {...props}
    />
  );
});
TooltipContent.displayName = "TooltipContent";

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
