import { cn } from "lib";
import type { Icon } from "./types";

export const ClockIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
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
      <circle cx="12" cy="12" r="10" />
      <path d="M12 8V12L14 14" />
    </svg>
  );
};
