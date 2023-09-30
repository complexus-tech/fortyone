import * as TooltipPrimitive from '@radix-ui/react-tooltip';
import { cn } from 'lib';
import { ReactElement, ReactNode } from 'react';

export const Tooltip = ({
  children,
  title,
  className = '',
}: {
  children: ReactElement;
  title: ReactNode;
  className?: string;
}) => {
  return (
    <TooltipPrimitive.Provider>
      <TooltipPrimitive.Root delayDuration={300}>
        <TooltipPrimitive.Trigger>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            className={cn(
              'dark:text-white z-50 text-blue-darker border bg-white/60 px-3 text-[0.9rem] py-3 dark:border-gray-300 font-semibold dark:bg-blue-darker/60 backdrop-blur rounded-xl',
              className
            )}
            sideOffset={3}
          >
            {title}
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  );
};
