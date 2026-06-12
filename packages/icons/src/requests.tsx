import { cn } from "lib";
import type { Icon } from "./types";

export const RequestsIcon = (props: Icon) => {
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
      <path d="M4 6.75C4 5.231 5.231 4 6.75 4h10.5C18.769 4 20 5.231 20 6.75v10.5C20 18.769 18.769 20 17.25 20H6.75C5.231 20 4 18.769 4 17.25V6.75Z" />
      <path d="M8 8.5h8" />
      <path d="M8 12h4.25" />
      <path d="M8 15.5h3" />
      <path d="m14.25 15.25 1.5 1.5L19 13.5" />
    </svg>
  );
};
