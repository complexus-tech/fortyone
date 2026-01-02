"use client";

import { ReactNode } from "react";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import { cn } from "lib";

type ContentProps = Omit<TooltipPrimitive.TooltipContentProps, "title"> & {
  title?: ReactNode;
};
export const Tooltip = ({
  children,
  title,
  className = "",
  sideOffset = 3,
  ...rest
}: ContentProps) => {
  if (!title) {
    return children;
  }
  return (
    <TooltipPrimitive.Provider>
      <TooltipPrimitive.Root delayDuration={600}>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            className={cn(
              "z-50 text-foreground border border-border bg-surface-elevated/90 px-3 text-[0.95rem] py-[0.35rem] font-medium backdrop-blur rounded-2xl mr-2",
              className
            )}
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
