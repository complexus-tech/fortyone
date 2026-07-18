"use client";

import { ReactNode } from "react";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import { cn } from "lib";

type ContentProps = Omit<TooltipPrimitive.TooltipContentProps, "title"> & {
  title?: ReactNode;
  delayDuration?: number;
};
export const Tooltip = ({
  children,
  title,
  className = "",
  collisionPadding = 12,
  sideOffset = 3,
  delayDuration = 600,
  ...rest
}: ContentProps) => {
  if (!title) {
    return children;
  }
  return (
    <TooltipPrimitive.Provider>
      <TooltipPrimitive.Root delayDuration={delayDuration}>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            className={cn(
              "z-50 mr-2 max-w-80 wrap-break-word whitespace-normal rounded-lg border-[0.5px] border-border/60 bg-surface-elevated px-3 py-2 text-left text-[0.95rem] leading-snug font-medium text-foreground shadow-lg shadow-shadow backdrop-blur-md dark:bg-surface-elevated/80",
              className,
            )}
            collisionPadding={collisionPadding}
            sideOffset={sideOffset}
            {...rest}
          >
            {title}
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  );
};
