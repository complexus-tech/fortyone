import { cn } from "lib";
import type { Icon } from "./types";

export const OKRIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="currentColor"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="12" cy="12" r="11" stroke="currentColor" fill="none" />
      <circle cx="12" cy="12" r="7" stroke="currentColor" fill="none" />
      <circle cx="12" cy="12" r="3" fill="currentColor" />
      <line x1="12" y1="1" x2="12" y2="3" stroke="currentColor" />
      <line x1="12" y1="21" x2="12" y2="23" stroke="currentColor" />
      <line x1="1" y1="12" x2="3" y2="12" stroke="currentColor" />
      <line x1="21" y1="12" x2="23" y2="12" stroke="currentColor" />
    </svg>
  );
};
