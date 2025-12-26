"use client";

import * as SelectPrimitive from "@radix-ui/react-select";
import { cn } from "lib";
import { ArrowDownIcon, ArrowUpIcon, CheckIcon } from "icons";
import { ComponentPropsWithoutRef, ElementRef, forwardRef } from "react";

type SelectProps = ComponentPropsWithoutRef<typeof SelectPrimitive.Root>;
export const Select = (props: SelectProps) => (
  <SelectPrimitive.Root {...props} />
);

const Group = SelectPrimitive.Group;

const Input = SelectPrimitive.Value;

const Trigger = forwardRef<
  ElementRef<typeof SelectPrimitive.Trigger>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Trigger
    ref={ref}
    className={cn(
      "flex h-[2.1rem] w-full items-center justify-between whitespace-nowrap text-gray dark:text-gray-200 outline-none rounded-[0.6rem] border border-gray-100/80 dark:border-dark-50 px-3 py-1 text-sm dark:bg-dark-100/30 bg-gray-50/30",
      className
    )}
    {...props}
  >
    {children}
    <SelectPrimitive.Icon asChild>
      <ArrowDownIcon className="ml-1 h-3.5" />
    </SelectPrimitive.Icon>
  </SelectPrimitive.Trigger>
));
Trigger.displayName = SelectPrimitive.Trigger.displayName;

const ScrollUpButton = forwardRef<
  ElementRef<typeof SelectPrimitive.ScrollUpButton>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollUpButton>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.ScrollUpButton
    ref={ref}
    className={cn(
      "flex cursor-default items-center justify-center py-1",
      className
    )}
    {...props}
  >
    <ArrowUpIcon className="h-4 w-auto" />
  </SelectPrimitive.ScrollUpButton>
));
ScrollUpButton.displayName = SelectPrimitive.ScrollUpButton.displayName;

const ScrollDownButton = forwardRef<
  ElementRef<typeof SelectPrimitive.ScrollDownButton>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollDownButton>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.ScrollDownButton
    ref={ref}
    className={cn(
      "flex cursor-default items-center justify-center py-1",
      className
    )}
    {...props}
  >
    <ArrowDownIcon className="h-4 w-auto" />
  </SelectPrimitive.ScrollDownButton>
));
ScrollDownButton.displayName = SelectPrimitive.ScrollDownButton.displayName;

const SelectContent = forwardRef<
  ElementRef<typeof SelectPrimitive.Content>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = "popper", ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Content
      ref={ref}
      className={cn(
        "relative z-50 max-h-96 min-w-32 overflow-hidden rounded-2xl border shadow-md bg-white/80 dark:bg-dark-200/90 dark:text-gray-200 backdrop-blur text-gray border-gray-100 dark:border-dark-50 w-max shadow-dark/10 dark:shadow-dark/20",
        position === "popper" &&
          "data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
        className
      )}
      position={position}
      {...props}
    >
      <ScrollUpButton />
      <SelectPrimitive.Viewport
        className={cn(
          "p-1",
          position === "popper" &&
            "h-(--radix-select-trigger-height) w-full min-w-(--radix-select-trigger-width)"
        )}
      >
        {children}
      </SelectPrimitive.Viewport>
      <ScrollDownButton />
    </SelectPrimitive.Content>
  </SelectPrimitive.Portal>
));
SelectContent.displayName = SelectPrimitive.Content.displayName;

const Label = forwardRef<
  ElementRef<typeof SelectPrimitive.Label>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.Label>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Label
    ref={ref}
    className={cn("px-2 py-1.5 text-sm font-semibold", className)}
    {...props}
  />
));
Label.displayName = SelectPrimitive.Label.displayName;

const Item = forwardRef<
  ElementRef<typeof SelectPrimitive.Item>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item
    ref={ref}
    className={cn(
      "relative flex w-full hover:dark:bg-dark-50/60 rounded-[0.6rem] hover:bg-gray-100/50 select-none items-center py-[0.3rem] pl-2 pr-8 text-sm outline-none data-disabled:pointer-events-none data-disabled:opacity-50",
      className
    )}
    {...props}
  >
    <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
      <SelectPrimitive.ItemIndicator>
        <CheckIcon className="h-4 w-auto" />
      </SelectPrimitive.ItemIndicator>
    </span>
    <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
));
Item.displayName = SelectPrimitive.Item.displayName;

const Separator = forwardRef<
  ElementRef<typeof SelectPrimitive.Separator>,
  ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Separator
    ref={ref}
    className={cn("-mx-1 my-1 h-px bg-primary", className)}
    {...props}
  />
));
Separator.displayName = SelectPrimitive.Separator.displayName;

Select.Trigger = Trigger;
Select.Content = SelectContent;
Select.Label = Label;
Select.Group = Group;
Select.Option = Item;
Select.Separator = Separator;
Select.Input = Input;
