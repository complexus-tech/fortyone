import { cn } from "lib";
import type { Icon } from "./types";

export const HttpIcon = (props: Icon) => {
  const { className, strokeWidth = 1.5, ...rest } = props;

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
      <path d="M2 7.75V16.25M5.5 7.75V16.25M2 12H5.5" />
      <path d="M7 7.75H11M9 7.75V16.25" />
      <path d="M12 7.75H16M14 7.75V16.25" />
      <path d="M17.5 16.25V7.75H19.5C20.8807 7.75 22 8.86929 22 10.25C22 11.6307 20.8807 12.75 19.5 12.75H17.5" />
    </svg>
  );
};
