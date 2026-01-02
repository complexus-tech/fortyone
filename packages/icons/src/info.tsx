import { cn } from "lib";
import type { Icon } from "./types";

export const InfoIcon = (props: Icon) => {
  const { className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M1.25 12C1.25 6.06294 6.06294 1.25 12 1.25C17.9371 1.25 22.75 6.06294 22.75 12C22.75 17.9371 17.9371 22.75 12 22.75C6.06294 22.75 1.25 17.9371 1.25 12ZM12 7.5C12.5523 7.5 13 7.94772 13 8.5V12.5C13 13.0523 12.5523 13.5 12 13.5C11.4477 13.5 11 13.0523 11 12.5V8.5C11 7.94772 11.4477 7.5 12 7.5ZM12 14.5C11.4477 14.5 11 14.9477 11 15.5C11 16.0523 11.4477 16.5 12 16.5H12.009C12.5613 16.5 13.009 16.0523 13.009 15.5C13.009 14.9477 12.5613 14.5 12.009 14.5H12Z"
        fill="currentColor"
      />
    </svg>
  );
};
