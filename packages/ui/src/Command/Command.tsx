"use client";

import { Command as CommandPrimitive } from "cmdk";

import { cn } from "lib";
import {
  ComponentProps,
  ComponentPropsWithoutRef,
  ElementRef,
  HTMLAttributes,
  forwardRef,
} from "react";
import { cva, VariantProps } from "cva";
import { SearchIcon } from "icons";

// type CommandProps = ComponentPropsWithoutRef<typeof CommandPrimitive>;

export const Command = ({ className, ...props }: any) => (
  <CommandPrimitive
    className={cn(
      "flex h-full w-full flex-col overflow-hidden rounded-md",
      className
    )}
    {...props}
  />
);
Command.displayName = CommandPrimitive.displayName;

const CommandInput = forwardRef<
  React.ElementRef<typeof CommandPrimitive.Input>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input>
>(({ className, ...props }, ref) => (
  <div className="flex items-center" cmdk-input-wrapper="">
    <SearchIcon className="h-[1.15rem] w-auto relative left-2.5 opacity-60" />
    <CommandPrimitive.Input
      ref={ref}
      className={cn(
        "bg-transparent placeholder:text-gray/80 placeholder:dark:text-gray-200/60 py-[0.15rem] pl-[1.1rem] outline-none w-full",
        className
      )}
      {...props}
    />
  </div>
));

CommandInput.displayName = CommandPrimitive.Input.displayName;

const contentClasses = cva(
  "bg-white/80 dark:bg-dark-200/80 backdrop-blur z-50 border border-gray-50 dark:border-dark-50/60 w-max shadow-sm shadow-dark/10 dark:shadow-dark/20 mt-1 py-1",
  {
    variants: {
      rounded: {
        sm: "rounded",
        md: "rounded-lg",
        lg: "rounded-xl",
      },
    },
    defaultVariants: {
      rounded: "md",
    },
  }
);

const CommandList = forwardRef<
  React.ElementRef<typeof CommandPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.List> &
    VariantProps<typeof contentClasses>
>(({ className, rounded, ...props }, ref) => (
  <CommandPrimitive.List
    ref={ref}
    className={cn(contentClasses({ rounded }), className)}
    {...props}
  />
));

CommandList.displayName = CommandPrimitive.List.displayName;

const CommandEmpty = forwardRef<
  React.ElementRef<typeof CommandPrimitive.Empty>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Empty>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Empty
    ref={ref}
    className={cn("py-4 text-center text-sm", className)}
    {...props}
  />
));

CommandEmpty.displayName = CommandPrimitive.Empty.displayName;

const CommandGroup = forwardRef<
  React.ElementRef<typeof CommandPrimitive.Group>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Group>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Group
    ref={ref}
    className={cn("px-1.5", className)}
    {...props}
  />
));

CommandGroup.displayName = CommandPrimitive.Group.displayName;

const CommandSeparator = forwardRef<
  ElementRef<typeof CommandPrimitive.Separator>,
  ComponentPropsWithoutRef<typeof CommandPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Separator
    ref={ref}
    className={cn(
      "border-t border-gray-100 dark:border-dark-100 my-2",
      className
    )}
    {...props}
  />
));
CommandSeparator.displayName = CommandPrimitive.Separator.displayName;

const CommandItem = forwardRef<
  ElementRef<typeof CommandPrimitive.Item>,
  ComponentPropsWithoutRef<typeof CommandPrimitive.Item> & { active?: boolean }
>(({ className, active, ...props }, ref) => (
  <CommandPrimitive.Item
    ref={ref}
    className={cn(
      "flex aria-selected:bg-gray-100/50 aria-selected:dark:bg-dark-50/35 gap-2 items-center select-none focus:dark:bg-dark-100/70 hover:dark:bg-dark-50 hover:bg-gray-100/70 focus:bg-gray-50 rounded-[0.4rem] w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
      {
        "bg-gray-100/70 dark:bg-dark-50/60": active,
      },
      className
    )}
    {...props}
  />
));

CommandItem.displayName = CommandPrimitive.Item.displayName;

const CommandShortcut = ({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) => {
  return (
    <span
      className={cn("ml-auto text-xs tracking-widest", className)}
      {...props}
    />
  );
};

Command.Input = CommandInput;
Command.List = CommandList;
Command.Empty = CommandEmpty;
Command.Group = CommandGroup;
Command.Item = CommandItem;
Command.Shortcut = CommandShortcut;
Command.Separator = CommandSeparator;
