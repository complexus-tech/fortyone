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
  "bg-white/80 dark:bg-dark-300/90 backdrop-blur z-50 border border-gray-50 dark:border-dark-100 w-max shadow shadow-gray-100 dark:shadow-dark/20 mt-1 py-1.5",
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
      "flex w-full cursor-pointer select-none items-center gap-1.5 rounded-[0.5rem] px-2 py-1.5 outline-none hover:bg-gray-100/70 focus:bg-gray-100/70 data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50 hover:dark:bg-dark-50 focus:dark:bg-dark-50",
      {
        "bg-gray-100/80 dark:bg-dark-50": active,
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
      "flex w-full cursor-pointer select-none items-center gap-1.5 rounded-lg px-2 py-1.5 outline-none hover:bg-gray-50 focus:bg-gray-50 data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[state=open]:bg-gray-50/80 data-[disabled]:opacity-50 hover:dark:bg-dark-50 focus:dark:bg-dark-50 data-[state=open]:dark:bg-dark-50",
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
      "my-1.5 border-b-[0.5px] border-gray-100 dark:border-dark-50",
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
