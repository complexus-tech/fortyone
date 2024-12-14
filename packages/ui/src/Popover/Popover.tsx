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
        "z-50 mt-1 mr-2 min-w-[12rem] rounded-[0.5rem] border border-gray-100 bg-white/80 py-2 text-dark shadow shadow-dark/10 backdrop-blur dark:border-dark-100 dark:bg-dark-200/80 dark:text-gray-200 dark:shadow-dark/20",
        className
      )}
      {...props}
    />
  </PopoverPrimitive.Portal>
));

PopoverContent.displayName = PopoverPrimitive.Content.displayName;

Popover.Trigger = PopoverTrigger;
Popover.Content = PopoverContent;
