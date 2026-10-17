import type { Icon } from "./types";
import { cn } from "lib";

export const TranscribeIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
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
      <path d="M11 9V17" />
      <path d="M15 15L15 19" />
      <path d="M7 3V21" />
      <path d="M3 9V15" />
      <path d="M17 4.5V11M17 11H15.5M17 11H18.5M21 5.5C21 4.67157 20.3284 4 19.5 4H14.5C13.6716 4 13 4.67157 13 5.5" />
    </svg>
  );
};
