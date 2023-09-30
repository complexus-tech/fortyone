import React from 'react';
import { ComponentProps } from 'react';
import * as TabsPrimitive from '@radix-ui/react-tabs';
import { cn } from 'lib';

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
        'relative w-max top-[1px] border-b-[3px] border-transparent dark:text-white px-1 pb-[12px] font-medium hover:text-primary-200 focus:text-primary focus:outline-0 data-[state=active]:border-b-primary dark:data-[state=active]:text-primary data-[state=active]:text-primary-200 flex items-center gap-2',
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
        'flex gap-4 flex-wrap border-b-[1.5px] border-gray-100 dark:border-gray-300 px-6 md:gap-8',
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
