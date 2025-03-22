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
              "dark:text-gray-200 z-50 text-gray border border-gray-100 bg-white/80 px-3 text-[0.95rem] py-[0.35rem] dark:border-dark-100/80 font-medium dark:bg-dark-200 backdrop-blur rounded-[0.6rem] mr-2",
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
