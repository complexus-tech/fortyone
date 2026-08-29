import { cn } from "lib";
import type { Icon } from "./types";

export const MoonIcon = (props: Icon) => {
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
      <path d="M14.25 11.15A6.15 6.15 0 0 1 6.85 3.75a5.7 5.7 0 1 0 7.4 7.4Z" />
    </svg>
  );
};
