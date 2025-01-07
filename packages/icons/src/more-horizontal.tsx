import { cn } from "lib";
import type { Icon } from "./types";

export const MoreHorizontalIcon = (props: Icon) => {
  const { strokeWidth = 4, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      strokeWidth={strokeWidth}
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M11.9959 12H12.0049"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M17.9998 12H18.0088"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M5.99981 12H6.00879"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
