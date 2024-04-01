"use client";
import * as ContextMenuPrimitive from "@radix-ui/react-context-menu";
import { VariantProps, cva } from "cva";
import {
  ComponentProps,
  ComponentPropsWithoutRef,
  ElementRef,
  forwardRef,
} from "react";

import { cn } from "lib";

type TriggerProps = ComponentProps<typeof ContextMenuPrimitive.Trigger>;
const Trigger = ({ children, className, ...rest }: TriggerProps) => (
  <ContextMenuPrimitive.Trigger
    className={cn("outline-none", className)}
    {...rest}
    asChild
  >
    {children}
  </ContextMenuPrimitive.Trigger>
);

const contentClasses = cva(
  "bg-white/80 dark:bg-dark-300/60 dark:text-gray-200 backdrop-blur text-gray z-50 border-[0.5px] border-gray-100 dark:border-dark-100 w-max shadow-lg shadow-dark/10 dark:shadow-dark/20 mt-1 py-2",
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

type ContentProps = ComponentProps<typeof ContextMenuPrimitive.Content> &
  VariantProps<typeof contentClasses>;

const Items = ({
  children,
  className,
  loop = true,
  rounded,
  ...rest
}: ContentProps) => {
  const classes = cn(contentClasses({ rounded }), className);
  return (
    <ContextMenuPrimitive.Portal>
      <ContextMenuPrimitive.Content className={classes} loop={loop} {...rest}>
        {children}
      </ContextMenuPrimitive.Content>
    </ContextMenuPrimitive.Portal>
  );
};

type SubContentProps = ComponentProps<typeof ContextMenuPrimitive.SubContent> &
  VariantProps<typeof contentClasses>;
const SubItems = ({
  children,
  className,
  sideOffset = 7.8,
  alignOffset = -4,
  loop = true,
  rounded,
  ...rest
}: SubContentProps) => {
  const classes = cn(contentClasses({ rounded }), className);
  return (
    <ContextMenuPrimitive.Portal>
      <ContextMenuPrimitive.SubContent
        className={classes}
        sideOffset={sideOffset}
        alignOffset={alignOffset}
        loop={loop}
        {...rest}
      >
        {children}
      </ContextMenuPrimitive.SubContent>
    </ContextMenuPrimitive.Portal>
  );
};

const Item = forwardRef<
  ElementRef<typeof ContextMenuPrimitive.Item>,
  ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Item> & {
    active?: boolean;
  }
>(({ children, className, active, ...rest }, ref) => (
  <ContextMenuPrimitive.Item
    className={cn(
      "flex gap-2 items-center select-none focus:dark:bg-dark-50 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-[0.45rem] w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
      {
        "bg-gray-50/80 dark:bg-dark-100": active,
      },
      className
    )}
    ref={ref}
    {...rest}
  >
    {children}
  </ContextMenuPrimitive.Item>
));

const SubTrigger = forwardRef<
  ElementRef<typeof ContextMenuPrimitive.SubTrigger>,
  ComponentPropsWithoutRef<typeof ContextMenuPrimitive.SubTrigger> & {
    active?: boolean;
  }
>(({ children, className, active, ...rest }, ref) => (
  <ContextMenuPrimitive.SubTrigger
    className={cn(
      "flex gap-2 items-center select-none data-[state=open]:bg-gray-50/80 data-[state=open]:dark:bg-dark-50 focus:dark:bg-dark-50 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-lg w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
      {
        "bg-gray-50/80 dark:bg-dark-100": active,
      },
      className
    )}
    ref={ref}
    {...rest}
  >
    {children}
  </ContextMenuPrimitive.SubTrigger>
));

type MenuProps = ComponentProps<typeof ContextMenuPrimitive.Root>;
export const ContextMenu = ({ children, ...rest }: MenuProps) => {
  return (
    <ContextMenuPrimitive.Root {...rest}>
      <div>{children}</div>
    </ContextMenuPrimitive.Root>
  );
};

const Separator = forwardRef<
  ElementRef<typeof ContextMenuPrimitive.Separator>,
  ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Separator>
>(({ className, ...rest }, ref) => (
  <ContextMenuPrimitive.Separator
    className={cn(
      "border-gray-100 dark:border-dark-100 border-b-[0.5px] my-3",
      className
    )}
    ref={ref}
    {...rest}
  />
));

const Group = forwardRef<
  ElementRef<typeof ContextMenuPrimitive.Group>,
  ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Group>
>(({ className, ...rest }, ref) => (
  <ContextMenuPrimitive.Group
    className={cn("px-2", className)}
    ref={ref}
    {...rest}
  />
));

ContextMenu.Trigger = Trigger;
ContextMenu.Separator = Separator;
ContextMenu.Group = Group;
ContextMenu.Items = Items;
ContextMenu.SubMenu = ContextMenuPrimitive.Sub;
ContextMenu.SubTrigger = SubTrigger;
ContextMenu.SubItems = SubItems;
ContextMenu.Item = Item;
