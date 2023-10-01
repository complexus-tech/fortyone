'use client';
import * as SwitchPrimitive from '@radix-ui/react-switch';
import { ComponentPropsWithoutRef } from 'react';

import { cn } from 'lib';

type SwitchProps = ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>;

export const Switch = ({ className = '', children, ...props }: SwitchProps) => {
  return (
    <SwitchPrimitive.Root
      className={cn(
        'w-[26px] h-[13px] bg-gray-200 dark:bg-dark rounded-full data-[state=checked]:bg-primary dark:data-[state=checked]:bg-primary transition',
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          'block h-[10px] translate-x-[2.5px] -translate-y-[0.2px] aspect-square bg-white rounded-full data-[state=checked]:translate-x-[14px] will-change-transform transition',
          className
        )}
      />
    </SwitchPrimitive.Root>
  );
};
