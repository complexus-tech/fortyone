import React from 'react';
import {
  DayPicker,
  DayPickerDefaultProps,
  DayPickerMultipleProps,
  DayPickerRangeProps,
  DayPickerSingleProps,
} from 'react-day-picker';
import { FiChevronLeft, FiChevronRight } from 'react-icons/fi';

import { cn } from '@/lib';

export type CalendarProps =
  | DayPickerDefaultProps
  | DayPickerSingleProps
  | DayPickerMultipleProps
  | DayPickerRangeProps;

export const Calendar = ({
  className,
  classNames,
  showOutsideDays = true,
  ...props
}: CalendarProps) => {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn(
        'px-4 py-5 rounded-xl border-[1.5px] border-gray-100 dark:border-gray-300 bg-white z-50 dark:bg-blue-dark w-80',
        className
      )}
      classNames={{
        months:
          'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0 dark:text-white text-lg',
        month: 'space-y-4',
        caption: 'flex justify-center pt-1 relative items-center',
        caption_label: 'text-base font-medium',
        nav: 'space-x-1 flex items-center',
        nav_button:
          'h-8 aspect-square border-[1.5px] border-gray-100 dark:border-gray-300 rounded-lg flex items-center justify-center',
        nav_button_previous: 'absolute left-1',
        nav_button_next: 'absolute right-1',
        table: 'w-full border-collapse space-y-1',
        head_row: 'flex',
        head_cell: 'w-10 font-semibold text-[1rem]',
        row: 'flex w-full mt-2',
        cell: 'text-center text-[1rem] p-0 relative focus-within:relative focus-within:z-50',
        day: 'h-10 w-10 p-0 font-normal flex items-center justify-center aria-selected:opacity-100 rounded-lg hover:bg-primary hover:text-white',
        day_selected: 'bg-primary text-primary rounded-lg',
        day_today:
          'bg-accent text-accent-foreground border border-gray-50 dark:border-gray-300/60',
        day_outside: 'opacity-50',
        day_disabled: 'opacity-50',
        day_range_middle: 'aria-selected:bg-primary aria-selected:text-white',
        day_hidden: 'invisible',
        ...classNames,
      }}
      components={{
        IconLeft: () => <FiChevronLeft className='h-5 w-auto' />,
        IconRight: () => <FiChevronRight className='h-5 w-auto' />,
      }}
      {...props}
    />
  );
};
