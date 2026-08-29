import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Time02Icon, Stroke Rounded, copied from
// @hugeicons/core-free-icons@4.2.3 (MIT license).
export const Time02Icon = (props: Icon) => {
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
        d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22Z"
        stroke="currentColor"
        strokeWidth={strokeWidth}
      />
      <path
        d="M12.0078 10.5082C11.1794 10.5082 10.5078 11.1798 10.5078 12.0082C10.5078 12.8366 11.1794 13.5082 12.0078 13.5082C12.8362 13.5082 13.5078 12.8366 13.5078 12.0082ZM12.0078 10.5082V6.99902M15.0147 15.0198L13.0661 13.0712"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={strokeWidth}
      />
    </svg>
  );
};
