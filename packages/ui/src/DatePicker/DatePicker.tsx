"use client";
import { cn } from "lib";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { ComponentPropsWithoutRef, ElementRef, forwardRef } from "react";
import { Calendar, CalendarProps } from "../Calendar/Calendar";

const Popover = PopoverPrimitive.Root;

const PopoverTrigger = PopoverPrimitive.Trigger;

const PopoverContent = forwardRef<
  ElementRef<typeof PopoverPrimitive.Content>,
  ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(({ className, align = "end", sideOffset = 4, ...props }, ref) => (
  <PopoverPrimitive.Portal>
    <PopoverPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={cn("z-50", className)}
      {...props}
    />
  </PopoverPrimitive.Portal>
));
PopoverContent.displayName = PopoverPrimitive.Content.displayName;

export type PopoverProps = ComponentPropsWithoutRef<typeof Popover>;

export const DatePicker = (props: PopoverProps) => {
  const { children } = props;
  return <Popover {...props}>{children}</Popover>;
};

type TriggerProps = ComponentPropsWithoutRef<typeof PopoverTrigger>;
const Trigger = ({ children, ...rest }: TriggerProps) => {
  return (
    <PopoverTrigger asChild {...rest}>
      {children}
    </PopoverTrigger>
  );
};

const CalendarPrimitive = (props: CalendarProps) => {
  return (
    <PopoverContent className="w-auto p-0 z-50">
      <Calendar {...props} />
    </PopoverContent>
  );
};

DatePicker.Trigger = Trigger;
DatePicker.Calendar = CalendarPrimitive;
