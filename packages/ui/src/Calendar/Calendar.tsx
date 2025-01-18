"use client";
import { DayPicker, PropsBase } from "react-day-picker";
import "react-day-picker/style.css";

import { cn } from "lib";

export type CalendarProps = PropsBase;

export const Calendar = ({
  className,
  classNames,
  showOutsideDays = true,
  ...props
}: CalendarProps) => {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      captionLayout="dropdown"
      mode="single"
      selected={new Date()}
      classNames={
        {
          // root: "px-4 py-5 rounded-xl border border-gray-100/90 dark:border-dark-50 bg-white/50 z-50 dark:bg-dark-200/80 backdrop-blur w-max shadow-lg",
          // today:
          //   "border border-gray-200 dark:border-dark-50 hover:border-primary",
          // selected: "bg-primary border-primary rounded-lg",
          // chevron: "h-4 w-auto",
          // day: "size-10 p-0 flex items-center justify-center rounded-lg hover:bg-primary hover:text-white cursor-pointer",
          // outside: "opacity-50",
          // disabled: "opacity-50",
          // range_start: "rounded-r-none aria-selected:text-white",
          // range_end: "rounded-l-none aria-selected:text-white",
          // range_middle:
          //   "aria-selected:bg-primary/20 dark:aria-selected:bg-primary/10 aria-selected:text-black dark:aria-selected:text-white rounded-none",
          // hidden: "invisible",
          // caption_label: "hidden",
          // nav: "hidden",
          // button_next: "hidden",
          // button_previous: "hidden",
          // dropdowns: "flex gap-2 w-full pt-1 relative items-center",
          // dropdown_root: "relative inline-flex items-center",
          // dropdown: "bg-transparent appearance-none w-full",
          // months:
          //   "flex flex-col sm:flex-row space-y-4 sm:space-y-0 dark:text-white text-lg",
          // month: "ml-0",
          // week: "flex w-full mt-2",
          // weekdays: "flex",
          // weekday: "size-10 font-semibold border-0 text-[1rem]",
          // month_grid: "w-full border-collapse space-y-1",
        }
      }
    />
  );
};
