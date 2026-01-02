"use client";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { cn } from "lib";
import React, { ComponentProps } from "react";

type TabProps = ComponentProps<typeof TabsPrimitive.Trigger> & {
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
};
const Trigger = ({
  children,
  value,
  className,
  leftIcon,
  rightIcon,
  ...rest
}: TabProps) => {
  return (
    <TabsPrimitive.Trigger
      value={value}
      className={cn(
        "relative w-max data-[state=active]:border data-[state=active]:border-border rounded-[0.7rem] text-text-secondary px-3 py-[0.3rem] hover:text-text-primary focus-visible:bg-state-hover focus-visible:outline-0 data-[state=active]:text-text-primary data-[state=active]:bg-surface-elevated flex items-center gap-2",
        className
      )}
      {...rest}
    >
      {leftIcon}
      {children}
      {rightIcon}
    </TabsPrimitive.Trigger>
  );
};

type ListProps = ComponentProps<typeof TabsPrimitive.List>;
const List = ({ children, className, ...rest }: ListProps) => {
  return (
    <TabsPrimitive.List
      className={cn(
        "flex flex-wrap w-max mx-5 md:mx-12 rounded-[0.7rem] bg-surface-muted",
        className
      )}
      {...rest}
    >
      {children}
    </TabsPrimitive.List>
  );
};

type ContentProps = ComponentProps<typeof TabsPrimitive.Content>;
const Panel = ({ children, value, ...rest }: ContentProps) => {
  return (
    <TabsPrimitive.Content value={value} {...rest}>
      {children}
    </TabsPrimitive.Content>
  );
};

type TabsProps = ComponentProps<typeof TabsPrimitive.Root>;
export const Tabs = ({ children, ...rest }: TabsProps) => {
  return <TabsPrimitive.Root {...rest}>{children}</TabsPrimitive.Root>;
};

Tabs.Tab = Trigger;
Tabs.List = List;
Tabs.Panel = Panel;
