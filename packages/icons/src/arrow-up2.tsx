import { cn } from "lib";
import type { Icon } from "./types";

export const ArrowUp2Icon = (props: Icon) => {
  const { className, strokeWidth = 3, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 15L12 9L6 15"
        stroke="currentColor"
        strokeMiterlimit="16"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
