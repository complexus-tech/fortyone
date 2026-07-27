import { cn } from "lib";
import type { Icon } from "./types";

export const ActiveSprintIcon = (props: Icon) => {
  const { strokeWidth = 2.5, className, ...rest } = props;

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
      <circle cx="12" cy="12" r="9.25" stroke="currentColor" />
      <path
        d="M8.5 12.25L10.75 14.5L15.5 9.75"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
