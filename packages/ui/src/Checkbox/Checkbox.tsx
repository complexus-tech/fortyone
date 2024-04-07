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
      "h-[1.15rem] w-[1.15rem] rounded border border-gray-200 dark:border-dark-50 data-[state=checked]:border-primary focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={cn("flex items-center justify-center text-white")}
    >
      <CheckIcon className="h-5 w-auto relative -top-[1px]" strokeWidth={2.5} />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;
