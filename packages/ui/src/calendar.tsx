"use client";
import {
  DayPicker,
  DayPickerDefaultProps,
  DayPickerMultipleProps,
  DayPickerRangeProps,
  DayPickerSingleProps,
} from "react-day-picker";

import { cn } from "lib";
import { ArrowLeft2Icon, ArrowRight2Icon } from "icons";

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
      className={cn("px-4 py-5 w-max shadow-lg rounded-2xl", className)}
      classNames={{
        months:
          "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0 dark:text-white text-lg",
        month: "space-y-4",
        caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-base font-medium",
        nav: "space-x-1 flex items-center",
        nav_button:
          "h-9 aspect-square rounded-[0.6rem] flex items-center justify-center hover:bg-gray-50 dark:hover:bg-dark-100",
        nav_button_previous: "absolute -left-1",
        nav_button_next: "absolute -right-1",
        table: "w-full border-collapse space-y-1",
        head_row: "flex",
        head_cell: "w-10 font-semibold border-0 text-base",
        row: "flex w-full mt-2",
        cell: "text-center text-base p-0 relative focus-within:relative focus-within:z-50",
        day: "h-10 w-10 p-0 font-normal flex items-center justify-center aria-selected:opacity-100 rounded-full hover:bg-primary hover:text-white cursor-pointer leading-10",
        day_selected: "bg-primary border-primary rounded-full",
        day_today:
          "border border-gray-200 dark:border-gray/50 hover:border-primary bg-gray-50 dark:bg-dark-50",
        day_outside: "opacity-40",
        day_disabled: "opacity-40",
        day_range_start: "rounded-r-none aria-selected:text-white",
        day_range_end: "rounded-l-none aria-selected:text-white",
        day_range_middle:
          "aria-selected:bg-primary/20 dark:aria-selected:bg-primary/10 aria-selected:text-black dark:aria-selected:text-white rounded-none",
        day_hidden: "invisible",
        dropdown:
          "appearance-none absolute top-0 bottom-0 w-full m-0 p-0 border-0 bg-transparent left-0 z-2",
        dropdown_year: "relative inline-flex items-center",
        dropdown_month: "relative inline-flex items-center",
        ...classNames,
      }}
      components={{
        IconLeft: () => (
          <ArrowLeft2Icon
            className="h-5 w-auto dark:text-white/80"
            strokeWidth={2.5}
          />
        ),
        IconRight: () => (
          <ArrowRight2Icon
            className="h-5 w-auto dark:text-white/80"
            strokeWidth={2.5}
          />
        ),
      }}
      {...props}
    />
  );
};
