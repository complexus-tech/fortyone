import { cn } from "lib";
import type { Icon } from "./types";

export const Calendar04Icon = (props: Icon) => {
  const { className, strokeWidth = 1.5, ...rest } = props;

  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M16 2V6M8 2V6" />
      <path d="M13 4H11C7.229 4 5.343 4 4.172 5.172C3 6.343 3 8.229 3 12V14C3 17.771 3 19.657 4.172 20.828C5.343 22 7.229 22 11 22H13C16.771 22 18.657 22 19.828 20.828C21 19.657 21 17.771 21 14V12C21 8.229 21 6.343 19.828 5.172C18.657 4 16.771 4 13 4Z" />
      <path d="M3 10H21" />
    </svg>
  );
};

export const CalendarIcon = Calendar04Icon;
