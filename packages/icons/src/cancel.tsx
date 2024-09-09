import { cn } from "lib";
import type { Icon } from "./types";

export const CancelIcon = (props: Icon) => {
  const { className, strokeWidth = 2 } = props;
  return (
    <svg
      {...props}
      fill="currentColor"
      className={cn("h-5 w-auto", className)}
      strokeWidth={strokeWidth}
      fillOpacity={0.1}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M14.9994 15L9 9M9.00064 15L15 9"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22C17.5228 22 22 17.5228 22 12Z"
        stroke="currentColor"
      />
    </svg>
  );
};
