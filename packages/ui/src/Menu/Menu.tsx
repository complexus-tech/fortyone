"use client";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { VariantProps, cva } from "cva";
import {
  ComponentProps,
  ComponentPropsWithoutRef,
  ElementRef,
  InputHTMLAttributes,
  forwardRef,
} from "react";

import { cn } from "lib";
import { CheckIcon, SearchIcon } from "icons";

type TriggerProps = ComponentProps<typeof DropdownMenu.Trigger>;
export const Trigger = ({ children, className, ...rest }: TriggerProps) => (
  <DropdownMenu.Trigger
    className={cn("outline-none", className)}
    {...rest}
    asChild
  >
    {children}
  </DropdownMenu.Trigger>
);

const contentClasses = cva(
  "bg-white dark:bg-dark-200 backdrop-blur z-50 border border-gray-100 dark:border-dark-50 w-max shadow-lg shadow-gray-100 dark:shadow-dark/20 mt-1 py-2",
  {
    variants: {
      rounded: {
        sm: "rounded",
        md: "rounded-lg",
        lg: "rounded-2xl",
      },
    },
    defaultVariants: {
      rounded: "lg",
    },
  }
);

const SubTrigger = forwardRef<
  ElementRef<typeof DropdownMenu.SubTrigger>,
  ComponentPropsWithoutRef<typeof DropdownMenu.SubTrigger> & {
    active?: boolean;
  }
>(({ children, className, active, ...rest }, ref) => (
  <DropdownMenu.SubTrigger
    className={cn(
      "flex w-full cursor-pointer select-none items-center gap-1.5 rounded-[0.6rem] px-2 py-1.5 outline-none hover:bg-gray-50 focus:bg-gray-50 data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[state=open]:bg-gray-50/80 data-[disabled]:opacity-50 hover:dark:bg-dark-50 focus:dark:bg-dark-50 data-[state=open]:dark:bg-dark-50",
      {
        "bg-gray-50/80 dark:bg-dark-100": active,
      },
      className
    )}
    ref={ref}
    {...rest}
  >
    {children}
  </DropdownMenu.SubTrigger>
));

type ContentProps = ComponentProps<typeof DropdownMenu.Content> &
  VariantProps<typeof contentClasses>;

export const Items = ({
  children,
  className,
  sideOffset = 4,
  loop = true,
  rounded,
  ...rest
}: ContentProps) => {
  const classes = cn(contentClasses({ rounded }), className);
  return (
    <DropdownMenu.Portal>
      <DropdownMenu.Content
        sideOffset={sideOffset}
        className={classes}
        loop={loop}
        {...rest}
      >
        {children}
      </DropdownMenu.Content>
    </DropdownMenu.Portal>
  );
};

const Item = forwardRef<
  ElementRef<typeof DropdownMenu.Item>,
  ComponentPropsWithoutRef<typeof DropdownMenu.Item> & {
    active?: boolean;
  }
>(({ children, className, active, ...rest }, ref) => (
  <DropdownMenu.Item
    className={cn(
      "flex w-full cursor-pointer select-none items-center gap-1.5 rounded-[0.6rem] px-2 py-1.5 outline-none hover:bg-gray-100/70 focus:bg-gray-100/70 data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50 hover:dark:bg-dark-50 focus:dark:bg-dark-50",
      {
        "bg-gray-100/80 dark:bg-dark-50": active,
      },
      className
    )}
    ref={ref}
    {...rest}
  >
    {children}
  </DropdownMenu.Item>
));

const CheckboxItem = forwardRef<
  ElementRef<typeof DropdownMenu.CheckboxItem>,
  ComponentPropsWithoutRef<typeof DropdownMenu.CheckboxItem>
>(({ children, className, checked, ...rest }, ref) => (
  <DropdownMenu.CheckboxItem
    className={cn(
      "mb-1 flex w-full cursor-pointer select-none items-center gap-1.5 rounded-lg px-2 py-1.5 outline-none hover:bg-gray-50 focus:bg-gray-50 data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50 hover:dark:bg-dark-50 focus:dark:bg-dark-50/80",
      {
        "bg-gray-50/80 dark:bg-dark-50/60": checked,
      },
      className
    )}
    ref={ref}
    checked={checked}
    {...rest}
  >
    <>
      <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
        <DropdownMenu.ItemIndicator>
          <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
        </DropdownMenu.ItemIndicator>
      </span>
      {children}
    </>
  </DropdownMenu.CheckboxItem>
));

type MenuProps = ComponentProps<typeof DropdownMenu.Root>;
export const Menu = ({ children, ...rest }: MenuProps) => {
  return (
    <DropdownMenu.Root {...rest}>
      <div>{children}</div>
    </DropdownMenu.Root>
  );
};

type SubContentProps = ComponentProps<typeof DropdownMenu.SubContent> &
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
    <DropdownMenu.Portal>
      <DropdownMenu.SubContent
        className={classes}
        sideOffset={sideOffset}
        alignOffset={alignOffset}
        loop={loop}
        {...rest}
      >
        {children}
      </DropdownMenu.SubContent>
    </DropdownMenu.Portal>
  );
};

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {}
const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...rest }, ref) => (
    <div className="flex items-center gap-1">
      <SearchIcon
        className="relative -left-1 h-[1.15rem] w-auto text-gray dark:text-gray-300"
        strokeWidth={2.5}
      />
      <input
        className={cn(
          "w-full bg-transparent py-[0.15rem] outline-none",
          className
        )}
        ref={ref}
        {...rest}
      />
    </div>
  )
);

const Separator = forwardRef<
  ElementRef<typeof DropdownMenu.Separator>,
  ComponentPropsWithoutRef<typeof DropdownMenu.Separator>
>(({ className, ...rest }, ref) => (
  <DropdownMenu.Separator
    className={cn(
      "my-1.5 border-b-[0.5px] border-gray-100 dark:border-dark-50",
      className
    )}
    ref={ref}
    {...rest}
  />
));

const Group = forwardRef<
  ElementRef<typeof DropdownMenu.Group>,
  ComponentPropsWithoutRef<typeof DropdownMenu.Group>
>(({ className, ...rest }, ref) => (
  <DropdownMenu.Group className={cn("px-1.5", className)} ref={ref} {...rest} />
));

Menu.Button = Trigger;
Menu.Separator = Separator;
Menu.SubMenu = DropdownMenu.Sub;
Menu.SubTrigger = SubTrigger;
Menu.SubItems = SubItems;
Menu.Group = Group;
Menu.Items = Items;
Menu.Item = Item;
Menu.Input = Input;
Menu.CheckboxItem = CheckboxItem;
