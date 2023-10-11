'use client';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { VariantProps, cva } from 'cva';
import {
  ComponentProps,
  ComponentPropsWithoutRef,
  ElementRef,
  forwardRef,
} from 'react';
import { BsCheckLg } from 'react-icons/bs';

import { cn } from 'lib';
import { TbCheck } from 'react-icons/tb';

type TriggerProps = ComponentProps<typeof DropdownMenu.Trigger>;
export const Trigger = ({ children, className, ...rest }: TriggerProps) => (
  <DropdownMenu.Trigger
    className={cn('outline-none', className)}
    {...rest}
    asChild
  >
    {children}
  </DropdownMenu.Trigger>
);

const contentClasses = cva(
  'bg-white dark:bg-dark-100 dark:text-gray-200 bg-opacity-80 dark:bg-opacity-50 backdrop-blur text-gray-300 z-50 border border-gray-100 dark:border-dark-100 w-max shadow-lg shadow-dark/10 dark:shadow-dark/20 mt-1 py-2',
  {
    variants: {
      rounded: {
        sm: 'rounded',
        md: 'rounded-lg',
        lg: 'rounded-xl',
      },
    },
    defaultVariants: {
      rounded: 'lg',
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
      'flex gap-2 mb-1 items-center select-none focus:dark:bg-dark-50/80 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-lg w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none',
      {
        'bg-gray-50/80 dark:bg-dark-50/60': active,
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
      'flex gap-2 mb-1 items-center select-none focus:dark:bg-dark-50/80 hover:dark:bg-dark-50 hover:bg-gray-50 focus:bg-gray-50 rounded-lg w-full py-1.5 px-2 outline-none cursor-pointer data-[disabled]:opacity-50 data-[disabled]:cursor-not-allowed data-[disabled]:pointer-events-none',
      {
        'bg-gray-50/80 dark:bg-dark-50/60': checked,
      },
      className
    )}
    ref={ref}
    checked={checked}
    {...rest}
  >
    <>
      <span className='absolute left-2 flex h-3.5 w-3.5 items-center justify-center'>
        <DropdownMenu.ItemIndicator>
          <BsCheckLg className='h-5 w-auto' />
          <TbCheck className='h-5 w-auto' strokeWidth={2.1} />
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

const Separator = forwardRef<
  ElementRef<typeof DropdownMenu.Separator>,
  ComponentPropsWithoutRef<typeof DropdownMenu.Separator>
>(({ className, ...rest }, ref) => (
  <DropdownMenu.Separator
    className={cn(
      'border-gray-100 dark:border-dark-100 border-b my-3',
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
  <DropdownMenu.Group className={cn('px-2', className)} ref={ref} {...rest} />
));

Menu.Button = Trigger;
Menu.Separator = Separator;
Menu.Group = Group;
Menu.Items = Items;
Menu.Item = Item;
Menu.CheckboxItem = CheckboxItem;
