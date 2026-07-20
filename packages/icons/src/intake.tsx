import { cn } from "lib";
import type { Icon } from "./types";

export const IntakeIcon = (props: Icon) => {
  const { className, strokeWidth = 2.2, ...rest } = props;

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
      <path d="M4 4.75h3.25A4.75 4.75 0 0 1 12 9.5v2.25" />
      <path d="M20 4.75h-3.25A4.75 4.75 0 0 0 12 9.5" />
      <path d="m9.75 9.75 2.25 2.25 2.25-2.25" />
      <path d="M5.5 13.75h13v3.5a2 2 0 0 1-2 2h-9a2 2 0 0 1-2-2v-3.5Z" />
      <path d="M9 16.5h6" />
    </svg>
  );
};
