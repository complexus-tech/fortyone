import { cn } from "lib";
import type { Icon } from "./types";

export const Refresh01Icon = (props: Icon) => {
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
      <path d="M20.488 15C19.252 18.496 15.919 21 12 21C7.029 21 3 16.971 3 12C3 7.029 7.029 3 12 3C15.729 3 18.929 5.268 20.294 8.5" />
      <path d="M15 9H18C19.414 9 20.121 9 20.561 8.561C21 8.121 21 7.414 21 6V3" />
    </svg>
  );
};

export const ReloadIcon = Refresh01Icon;
