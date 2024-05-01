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
  "bg-white/80 dark:bg-dark-300/60 backdrop-blur z-50 border-[0.5px] border-gray-50 dark:border-dark-100 w-max shadow-sm shadow-dark/10 dark:shadow-dark/20 mt-1 py-1",
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
      "flex gap-2 mb-1 items-center select-none focus:dark:bg-dark-50 hover:dark:bg-dark-50 hover:bg-gray-100/50 focus:bg-gray-100/50 rounded-[0.4rem] w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
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
      "flex gap-2 mb-1 items-center select-none focus:dark:bg-dark-50/80 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-lg w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none",
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

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {}
const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...rest }, ref) => (
    <div className="flex items-center gap-1">
      <SearchIcon
        className="h-[1.15rem] w-auto relative -left-1 text-gray dark:text-gray-300"
        strokeWidth={2.5}
      />
      <input
        className={cn(
          "bg-transparent py-[0.15rem] outline-none w-full",
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
      "border-gray-100 dark:border-dark-100 border-b-[0.5px] my-2",
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
Menu.Group = Group;
Menu.Items = Items;
Menu.Item = Item;
Menu.Input = Input;
Menu.CheckboxItem = CheckboxItem;
