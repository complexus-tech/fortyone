import { cn } from "lib";
import type { Icon } from "./types";

export const SunIcon = (props: Icon) => {
  const { className, strokeWidth = 1.25, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="18"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 18 18"
      width="18"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="9" cy="9" r="2.75" />
      <path d="M9 1.75v1.5M9 14.75v1.5M16.25 9h-1.5M3.25 9h-1.5M14.13 3.87l-1.06 1.06M4.93 13.07l-1.06 1.06M14.13 14.13l-1.06-1.06M4.93 4.93 3.87 3.87" />
    </svg>
  );
};
