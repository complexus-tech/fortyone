import { cn } from "lib";
import type { Icon } from "./types";

export const ObjectiveIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 11L20.3458 8.84853C20.7819 8.44853 21 8.24853 21 8M18 5L20.3458 7.15147C20.7819 7.55147 21 7.75147 21 8M21 8C3 8 3 21 3 21"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="5.5" cy="5.5" r="2.5" />
      <path d="M13 21L18 16M18 21L13 16" strokeLinecap="round" />
    </svg>
  );
};
