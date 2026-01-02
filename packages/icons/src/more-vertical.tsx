import type { Icon } from "./types";
import { cn } from "lib";

export const MoreVerticalIcon = (props: Icon) => {
  const { className, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      className={cn("h-5 w-auto text-icon", className)}
    >
      <path
        d="M11.992 12H12.001"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M11.9842 18H11.9932"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M11.9998 6H12.0088"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
