import { cn } from "lib";
import type { Icon } from "./types";

export const MinusIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="none"
      height="24"
      strokeWidth={strokeWidth}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M20 12L4 12"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
