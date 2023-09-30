import { cn } from '@/lib';
import * as PopoverPrimitive from '@radix-ui/react-popover';
import { format } from 'date-fns';
import React, { ComponentPropsWithoutRef, ElementRef, forwardRef } from 'react';
import { AiOutlineCalendar } from 'react-icons/ai';
import { Calendar, CalendarProps } from '../Calendar/Calendar';

const Popover = PopoverPrimitive.Root;

const PopoverTrigger = PopoverPrimitive.Trigger;

const PopoverContent = forwardRef<
  ElementRef<typeof PopoverPrimitive.Content>,
  ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(({ className, align = 'end', sideOffset = 4, ...props }, ref) => (
  <PopoverPrimitive.Portal>
    <PopoverPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={cn('z-50', className)}
      {...props}
    />
  </PopoverPrimitive.Portal>
));
PopoverContent.displayName = PopoverPrimitive.Content.displayName;

export type DatePickerProps = CalendarProps & {
  placeholder?: string;
  label?: string;
  helpText?: string;
  required?: boolean;
};

export const DatePicker = (props: DatePickerProps) => {
  const {
    className,
    label,
    helpText,
    required,
    placeholder = 'Pick a date',
    ...calendarProps
  } = props;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <label className='relative block'>
          {label && (
            <span className='mb-3 block font-medium dark:text-white'>
              {label}
              {required && <span className='text-danger'>*</span>}
            </span>
          )}
          <button
            className={cn(
              'w-full rounded-xl border-[1.5px] text-left bg-transparent dark:bg-transparent border-gray-100 dark:border-gray-300 dark:ring-offset-secondary py-[0.95rem] px-5 dark:text-white flex items-center gap-3 justify-between focus:ring ring-primary',
              className
            )}
            type='button'
          >
            {props.selected ? (
              format(props.selected as Date, 'PPP')
            ) : (
              <span className='opacity-80'>{placeholder}</span>
            )}
            <AiOutlineCalendar className='h-5 w-auto relative -top-[0.5px]' />
          </button>
          {helpText && (
            <span className='text-[0.8rem] font-medium inline-block left-[2px] -bottom-5 absolute text-gray-300 first-letter:uppercase'>
              {helpText}
            </span>
          )}
        </label>
      </PopoverTrigger>
      <PopoverContent className='w-auto p-0 z-50'>
        <Calendar {...calendarProps} />
      </PopoverContent>
    </Popover>
  );
};
