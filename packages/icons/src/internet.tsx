import { cn } from "lib";
import type { Icon } from "./types";

export const InternetIcon = (props: Icon) => {
  const { className, strokeWidth = 2.3, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("w-auto h-5 text-icon", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      strokeWidth={strokeWidth}
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="12" cy="12" r="10" stroke="currentColor" />
      <ellipse cx="12" cy="12" rx="4" ry="10" stroke="currentColor" />
      <path
        d="M2 12H22"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
