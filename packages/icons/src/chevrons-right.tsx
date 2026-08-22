import { cn } from "lib";
import type { Icon } from "./types";

export const ChevronsRightIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;

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
      <path d="M6.00004 17C6.00004 17 11 13.3176 11 12C11 10.6824 6 7 6 7" />
      <path d="M13 17C13 17 18 13.3176 18 12C18 10.6824 13 7 13 7" />
    </svg>
  );
};
