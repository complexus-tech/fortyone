import { cn } from "lib";
import type { Icon } from "./types";

export const CalendarPlusIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="none"
      height="24"
      className={cn(className, "h-5 w-auto text-gray dark:text-gray-300")}
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 2V4M6 2V4"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M11.05 22C7.01949 22 5.00424 22 3.75212 20.6464C2.5 19.2927 2.5 17.1141 2.5 12.7568V12.2432C2.5 7.88594 2.5 5.70728 3.75212 4.35364C5.00424 3 7.01949 3 11.05 3H12.95C16.9805 3 18.9958 3 20.2479 4.35364C21.4765 5.68186 21.4996 7.80438 21.5 12"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M3 8H21"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M13 18H21M17 14L17 22"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
