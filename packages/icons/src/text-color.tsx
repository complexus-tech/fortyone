import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons TextColorIcon, Stroke Rounded, copied from
// @hugeicons/core-free-icons@4.3.4 (MIT license).
export const TextColorIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M3 21H21"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth={strokeWidth}
      />
      <path
        d="M19 18L15.6247 9.15847C14.0574 5.05282 13.2737 3 12 3C10.7263 3 9.94261 5.05282 8.37527 9.15847L5 18"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth={strokeWidth}
      />
      <path
        d="M8 11H16"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth={strokeWidth}
      />
    </svg>
  );
};
