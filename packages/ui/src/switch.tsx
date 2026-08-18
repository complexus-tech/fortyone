"use client";
import * as SwitchPrimitive from "@radix-ui/react-switch";
import { ComponentPropsWithoutRef } from "react";

import { cn } from "lib";

type SwitchProps = ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>;

export const Switch = ({ className, children, ...props }: SwitchProps) => {
  return (
    <SwitchPrimitive.Root
      className={cn(
        "relative flex h-4 min-h-4 w-7 min-w-7 shrink-0 flex-none items-center rounded-full border border-surface-prominent bg-surface-prominent transition dark:border-state-selected dark:bg-state-selected data-[state=checked]:border-primary data-[state=checked]:bg-primary",
        className,
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb className="pointer-events-none block size-3 shrink-0 translate-x-0.5 rounded-full bg-white shadow-sm transition-transform data-[state=checked]:translate-x-3" />
    </SwitchPrimitive.Root>
  );
};
