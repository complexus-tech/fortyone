"use client";
import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import { CheckIcon } from "icons";
import { cn } from "lib";
import { forwardRef } from "react";

export const Checkbox = forwardRef<
  React.ElementRef<typeof CheckboxPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>
>(({ className, ...props }, ref) => (
  <CheckboxPrimitive.Root
    ref={ref}
    className={cn(
      "h-[1.15rem] w-[1.15rem] rounded border border-gray-200 focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:border-primary dark:data-[state=checked]:border-primary data-[state=checked]:bg-primary dark:border-dark-50",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={cn("flex items-center justify-center text-white")}
    >
      <CheckIcon
        className="relative -top-[1.5px] text-current dark:text-current"
        strokeWidth={3}
      />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;
