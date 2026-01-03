"use client";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { cn } from "lib";
import { forwardRef, ElementRef, ComponentPropsWithoutRef } from "react";

type PopoverProps = ComponentPropsWithoutRef<typeof PopoverPrimitive.Root>;
export const Popover = (props: PopoverProps) => (
  <PopoverPrimitive.Root {...props} />
);

const PopoverTrigger = PopoverPrimitive.Trigger;

const PopoverContent = forwardRef<
  ElementRef<typeof PopoverPrimitive.Content>,
  ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(({ className, align = "center", sideOffset = 4, ...props }, ref) => (
  <PopoverPrimitive.Portal>
    <PopoverPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={cn(
        "z-50 mt-1 mr-2 min-w-48 rounded-2xl border py-2 text-foreground shadow shadow-shadow backdrop-blur-md border-border bg-surface-elevated/90",
        className
      )}
      {...props}
    />
  </PopoverPrimitive.Portal>
));

PopoverContent.displayName = PopoverPrimitive.Content.displayName;

Popover.Trigger = PopoverTrigger;
Popover.Content = PopoverContent;
