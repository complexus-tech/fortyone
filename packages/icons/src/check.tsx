import { cn } from "lib";
import type { Icon } from "./types";

export const CheckIcon = (props: Icon) => {
  const { strokeWidth = 2.3, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto", className)}
      fill="none"
      height="24"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5 14L8.5 17.5L19 6.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
