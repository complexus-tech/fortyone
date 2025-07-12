"use client";
import * as NavigationMenuPrimitive from "@radix-ui/react-navigation-menu";
import { cva } from "cva";
import { cn } from "lib";
import { ArrowDown2Icon } from "icons";
import { forwardRef, ElementRef, ComponentPropsWithoutRef } from "react";

const NavigationMenuViewport = forwardRef<
  ElementRef<typeof NavigationMenuPrimitive.Viewport>,
  ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.Viewport>
>(({ className, ...props }, ref) => (
  <div className={cn("absolute top-full flex justify-center", className)}>
    <NavigationMenuPrimitive.Viewport
      className={cn(
        "relative mt-3.5 h-[var(--radix-navigation-menu-viewport-height)] w-full overflow-hidden rounded-2xl dark:bg-dark-300 bg-white backdrop-blur border border-gray-100 dark:border-dark-50 shadow-lg md:w-[var(--radix-navigation-menu-viewport-width)]",
        className
      )}
      ref={ref}
      {...props}
    />
  </div>
));
NavigationMenuViewport.displayName =
  NavigationMenuPrimitive.Viewport.displayName;

type NavigationMenu = ComponentPropsWithoutRef<
  typeof NavigationMenuPrimitive.Root
> & {
  align?: "start" | "end";
};
export const NavigationMenu = ({
  className,
  children,
  align = "start",
  ...props
}: NavigationMenu) => (
  <NavigationMenuPrimitive.Root
    className={cn(
      "relative z-10 flex max-w-max flex-1 items-center justify-center",
      className
    )}
    {...props}
  >
    {children}
    <NavigationMenuViewport
      className={cn({
        "left-0": align === "start",
        "right-0": align === "end",
      })}
    />
  </NavigationMenuPrimitive.Root>
);

const NavigationMenuList = forwardRef<
  ElementRef<typeof NavigationMenuPrimitive.List>,
  ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.List>
>(({ className, ...props }, ref) => (
  <NavigationMenuPrimitive.List
    ref={ref}
    className={cn(
      "group flex flex-1 list-none items-center justify-center space-x-1",
      className
    )}
    {...props}
  />
));
NavigationMenuList.displayName = NavigationMenuPrimitive.List.displayName;

const NavigationMenuItem = NavigationMenuPrimitive.Item;

const navigationMenuTriggerStyle = cva(
  "group inline-flex gap-1 w-max items-center justify-center text-[0.95rem]"
);

const NavigationMenuTrigger = forwardRef<
  ElementRef<typeof NavigationMenuPrimitive.Trigger>,
  ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.Trigger> & {
    hideArrow?: boolean;
  }
>(({ className, children, hideArrow, ...props }, ref) => (
  <NavigationMenuPrimitive.Trigger
    ref={ref}
    className={cn(navigationMenuTriggerStyle(), "group", className)}
    {...props}
  >
    {children}
    {!hideArrow && (
      <ArrowDown2Icon
        strokeWidth={3}
        className="h-3 w-auto ease-linear transition duration-200 group-data-[state=open]:rotate-180"
        aria-hidden="true"
      />
    )}
  </NavigationMenuPrimitive.Trigger>
));
NavigationMenuTrigger.displayName = NavigationMenuPrimitive.Trigger.displayName;

const NavigationMenuContent = forwardRef<
  ElementRef<typeof NavigationMenuPrimitive.Content>,
  ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.Content>
>(({ className, ...props }, ref) => (
  <NavigationMenuPrimitive.Content
    ref={ref}
    className={cn("left-0 top-0 w-full md:absolute md:w-auto", className)}
    {...props}
  />
));
NavigationMenuContent.displayName = NavigationMenuPrimitive.Content.displayName;

const NavigationMenuLink = NavigationMenuPrimitive.Link;

const NavigationMenuIndicator = forwardRef<
  ElementRef<typeof NavigationMenuPrimitive.Indicator>,
  ComponentPropsWithoutRef<typeof NavigationMenuPrimitive.Indicator>
>(({ className, ...props }, ref) => (
  <NavigationMenuPrimitive.Indicator
    ref={ref}
    className={cn(
      "top-full z-[1] flex h-1.5 items-end justify-center overflow-hidden data-[state=visible]:animate-in data-[state=hidden]:animate-out data-[state=hidden]:fade-out data-[state=visible]:fade-in",
      className
    )}
    {...props}
  >
    <div className="relative top-[60%] h-2 w-2 rotate-45 rounded-tl-sm bg-border shadow-md" />
  </NavigationMenuPrimitive.Indicator>
));
NavigationMenuIndicator.displayName =
  NavigationMenuPrimitive.Indicator.displayName;

// export {
//   navigationMenuTriggerStyle,
// };

NavigationMenu.List = NavigationMenuList;
NavigationMenu.Item = NavigationMenuItem;
NavigationMenu.Content = NavigationMenuContent;
NavigationMenu.Trigger = NavigationMenuTrigger;
NavigationMenu.Link = NavigationMenuLink;
NavigationMenu.Indicator = NavigationMenuIndicator;
NavigationMenu.Viewport = NavigationMenuViewport;
