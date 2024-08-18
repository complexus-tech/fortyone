"use client";
import {
  DayPicker,
  DayPickerDefaultProps,
  DayPickerMultipleProps,
  DayPickerRangeProps,
  DayPickerSingleProps,
} from "react-day-picker";

import { cn } from "lib";
import { ArrowLeftIcon, ArrowRightIcon } from "icons";

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
        "px-4 py-5 rounded-xl border border-gray-100/90 dark:border-dark-50 bg-white/50 z-50 dark:bg-dark-200/80 backdrop-blur w-max shadow-lg",
        className
      )}
      classNames={{
        months:
          "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0 dark:text-white text-lg",
        month: "space-y-4",
        caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-base font-medium",
        nav: "space-x-1 flex items-center",
        nav_button:
          "h-8 aspect-square border border-gray-100 dark:border-dark-50 rounded-lg flex items-center justify-center",
        nav_button_previous: "absolute left-1",
        nav_button_next: "absolute right-1",
        table: "w-full border-collapse space-y-1",
        head_row: "flex",
        head_cell: "w-10 font-semibold border-0 text-[1rem]",
        row: "flex w-full mt-2",
        cell: "text-center text-[1rem] p-0 relative focus-within:relative focus-within:z-50",
        day: "h-10 w-10 p-0 font-normal flex items-center justify-center aria-selected:opacity-100 rounded-lg hover:bg-primary hover:text-white cursor-pointer",
        day_selected: "bg-primary border-primary rounded-lg",
        day_today:
          "border border-gray-200 dark:border-dark-50 hover:border-primary",
        day_outside: "opacity-50",
        day_disabled: "opacity-50",
        day_range_start: "rounded-r-none aria-selected:text-white",
        day_range_end: "rounded-l-none aria-selected:text-white",
        day_range_middle:
          "aria-selected:bg-primary/20 dark:aria-selected:bg-primary/10 aria-selected:text-black dark:aria-selected:text-white rounded-none",
        day_hidden: "invisible",
        dropdown:
          "appearance-none absolute top-0 bottom-0 w-full m-0 p-0 border-0 bg-transparent left-0 z-[2]",
        dropdown_year: "relative inline-flex items-center",
        dropdown_month: "relative inline-flex items-center",
        ...classNames,
      }}
      components={{
        IconLeft: () => <ArrowLeftIcon className="h-4 w-auto" />,
        IconRight: () => <ArrowRightIcon className="h-4 w-auto" />,
      }}
      {...props}
    />
  );
};
