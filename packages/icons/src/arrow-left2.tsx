import { cn } from "lib";
import type { Icon } from "./types";

export const ArrowLeft2Icon = (props: Icon) => {
  const { className, strokeWidth = 3, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M15 6L9 12.0001L15 18"
        stroke="currentColor"
        strokeMiterlimit="16"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
