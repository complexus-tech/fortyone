import { cn } from "lib";
import type { Icon } from "./types";

export const ExternalLinkIcon = (props: Icon) => {
  const { className, strokeWidth = 1.25, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-4 w-auto text-icon", className)}
      fill="none"
      height="16"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 16 16"
      width="16"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M4 12 12 4M6 4h6v6" />
    </svg>
  );
};
