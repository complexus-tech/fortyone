import { cn } from "lib";
import type { Icon } from "./types";

export const ChevronsLeftIcon = (props: Icon) => {
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
      <path d="M18 17C18 17 13 13.3176 13 12C13 10.6824 18 7 18 7" />
      <path d="M11 17C11 17 6.00001 13.3176 6 12C5.99999 10.6824 11 7 11 7" />
    </svg>
  );
};
