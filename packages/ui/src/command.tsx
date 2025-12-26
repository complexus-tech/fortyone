"use client";

import { Command as CommandPrimitive } from "cmdk";

import { cn } from "lib";
import { ComponentPropsWithoutRef, HTMLAttributes } from "react";
import { cva, VariantProps } from "cva";
import { SearchIcon } from "icons";

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

const CommandInput = ({
  className,
  icon = (
    <SearchIcon className="h-[1.15rem] w-auto relative left-2.5 opacity-60" />
  ),
  ...props
}: React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input> & {
  icon?: React.ReactNode;
}) => (
  <div className="flex items-center" cmdk-input-wrapper="">
    {icon}
    <CommandPrimitive.Input
      className={cn(
        "bg-transparent placeholder:text-gray/80 placeholder:dark:text-gray-200/60 py-[0.15rem] pl-[1.1rem] outline-none w-full",
        className
      )}
      {...props}
    />
  </div>
);

CommandInput.displayName = CommandPrimitive.Input.displayName;

const contentClasses = cva(
  "bg-white/70 dark:bg-dark-200/80 backdrop-blur-[12px] z-50 border border-gray-100 dark:border-dark-50 w-max shadow-xl shadow-gray-100 dark:shadow-dark/20 mt-1 py-1.5",
  {
    variants: {
      rounded: {
        sm: "rounded",
        md: "rounded-[0.6rem]",
        lg: "rounded-xl",
      },
    },
    defaultVariants: {
      rounded: "lg",
    },
  }
);

const CommandList = ({
  className,
  rounded,
  ...props
}: React.ComponentPropsWithoutRef<typeof CommandPrimitive.List> &
  VariantProps<typeof contentClasses>) => (
  <CommandPrimitive.List
    className={cn(contentClasses({ rounded }), className)}
    {...props}
  />
);

CommandList.displayName = CommandPrimitive.List.displayName;

const CommandEmpty = ({
  className,
  ...props
}: React.ComponentPropsWithoutRef<typeof CommandPrimitive.Empty>) => (
  <CommandPrimitive.Empty
    className={cn("py-4 text-center text-sm", className)}
    {...props}
  />
);

CommandEmpty.displayName = CommandPrimitive.Empty.displayName;

const CommandGroup = ({
  className,
  ...props
}: React.ComponentPropsWithoutRef<typeof CommandPrimitive.Group>) => (
  <CommandPrimitive.Group className={cn("px-1.5", className)} {...props} />
);

CommandGroup.displayName = CommandPrimitive.Group.displayName;

const CommandSeparator = ({
  className,
  ...props
}: ComponentPropsWithoutRef<typeof CommandPrimitive.Separator>) => (
  <CommandPrimitive.Separator
    className={cn(
      "border-t border-gray-100 dark:border-dark-100 my-2",
      className
    )}
    {...props}
  />
);
CommandSeparator.displayName = CommandPrimitive.Separator.displayName;

const CommandItem = ({
  className,
  active,
  ...props
}: ComponentPropsWithoutRef<typeof CommandPrimitive.Item> & {
  active?: boolean;
}) => (
  <CommandPrimitive.Item
    className={cn(
      "flex aria-selected:bg-gray-100/50 aria-selected:dark:bg-dark-50/50 gap-2 items-center select-none focus:dark:bg-dark-100/70 hover:dark:bg-dark-50 hover:bg-gray-100/70 focus:bg-gray-50 rounded-[0.6rem] w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
      {
        "bg-gray-100/70 dark:bg-dark-50": active,
      },
      className
    )}
    {...props}
  />
);

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
Command.Loading = CommandPrimitive.Loading;
