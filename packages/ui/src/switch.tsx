"use client";
import * as SwitchPrimitive from "@radix-ui/react-switch";
import { ComponentPropsWithoutRef } from "react";

import { cn } from "lib";

type SwitchProps = ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>;

export const Switch = ({ className, children, ...props }: SwitchProps) => {
  return (
    <SwitchPrimitive.Root
      className={cn(
        "w-[24px] h-[14px] border bg-border-strong/80 border-border-strong/80 data-[state=checked]:border-primary rounded-full data-[state=checked]:bg-primary transition",
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          "block h-[10px] translate-x-[1.5px] -translate-y-[0.2px] aspect-square bg-surface rounded-full data-[state=checked]:translate-x-[11px] will-change-transform transition",
          className
        )}
      />
    </SwitchPrimitive.Root>
  );
};
