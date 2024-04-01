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
        "z-50 min-w-[12rem] rounded-[0.6rem] bg-white/80 dark:bg-dark-300/80 dark:text-gray-200 backdrop-blur text-gray-300 border-[0.5px] border-gray-100 dark:border-dark-100 shadow shadow-dark/10 dark:shadow-dark/20 mt-1 py-2",
        className
      )}
      {...props}
    />
  </PopoverPrimitive.Portal>
));

PopoverContent.displayName = PopoverPrimitive.Content.displayName;

Popover.Trigger = PopoverTrigger;
Popover.Content = PopoverContent;
