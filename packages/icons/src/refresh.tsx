import { cn } from "lib";
import type { Icon } from "./types";

export const RefreshIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;

  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M21 12A9 9 0 1 1 18.36 5.64L21 8"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M21 3V8H16"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
