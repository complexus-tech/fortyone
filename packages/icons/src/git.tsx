import { cn } from "lib";
import type { Icon } from "./types";

export const GitIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
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
      <circle
        cx="6"
        cy="6"
        r="2.25"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle
        cx="6"
        cy="18"
        r="2.25"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle
        cx="18"
        cy="18"
        r="2.25"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6 8.25V15.75"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M13 6H15C16.6569 6 18 7.34315 18 9V15.75"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M14 3.75L11.75 6L14 8.25"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
