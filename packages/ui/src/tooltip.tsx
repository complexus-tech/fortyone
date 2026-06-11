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
              "z-50 mr-2 max-w-[calc(100vw-1.5rem)] rounded-2xl border border-border bg-surface-elevated/90 px-3 py-[0.35rem] text-[0.95rem] font-medium text-foreground backdrop-blur",
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
