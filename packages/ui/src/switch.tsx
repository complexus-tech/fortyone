"use client";
import * as SwitchPrimitive from "@radix-ui/react-switch";
import { ComponentPropsWithoutRef } from "react";

import { cn } from "lib";

type SwitchProps = ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>;

export const Switch = ({ className, children, ...props }: SwitchProps) => {
  return (
    <SwitchPrimitive.Root
      className={cn(
        "relative flex h-4 w-7 shrink-0 items-center rounded-full border border-surface-prominent bg-surface-prominent transition data-[state=checked]:border-primary data-[state=checked]:bg-primary",
        className,
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb className="pointer-events-none block size-3 translate-x-0.5 rounded-full bg-white shadow-sm transition-transform data-[state=checked]:translate-x-[0.875rem]" />
    </SwitchPrimitive.Root>
  );
};
