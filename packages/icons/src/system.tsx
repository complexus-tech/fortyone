import { cn } from "lib";
import type { Icon } from "./types";

export const SystemIcon = (props: Icon) => {
  const { className, strokeWidth = 1.25, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="18"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 18 18"
      width="18"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect height="8.5" rx="1.25" width="12.5" x="2.75" y="3.5" />
      <path d="M7 14.5h4M9 12v2.5" />
    </svg>
  );
};
