import { cn } from "lib";
import type { Icon } from "./types";

export const UnorderedListIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M8 5L20 5" stroke="currentColor" strokeLinecap="round" />
      <path
        d="M4 5H4.00898"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 12H4.00898"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 19H4.00898"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M8 12L20 12" stroke="currentColor" strokeLinecap="round" />
      <path d="M8 19L20 19" stroke="currentColor" strokeLinecap="round" />
    </svg>
  );
};
