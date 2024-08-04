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
  return (
    <TooltipPrimitive.Provider>
      <TooltipPrimitive.Root delayDuration={500}>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            className={cn(
              "dark:text-gray-200 z-50 text-gray border border-gray-100 bg-white/80 px-3 text-sm py-[0.35rem] dark:border-dark-100/60 font-medium dark:bg-dark-200/80 backdrop-blur rounded-lg",
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
