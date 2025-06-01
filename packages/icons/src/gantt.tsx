import { cn } from "lib";
import type { Icon } from "./types";

export const GanttIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect width="24" height="24" fill="none" />
      <rect x="2" y="5" width="10" height="4" fill="currentColor" rx="2" />
      <rect x="6" y="11" width="14" height="4" fill="currentColor" rx="2" />
      <rect x="4" y="17" width="10" height="4" fill="currentColor" rx="2" />
    </svg>
  );
};
