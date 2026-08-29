import { cn } from "lib";
import type { Icon } from "./types";

export const EnterIcon = (props: Icon) => {
  const { className, strokeWidth = 3, ...rest } = props;

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
      <path d="M9 10L4 15L9 20" />
      <path d="M20 4V11C20 13.2091 18.2091 15 16 15H4" />
    </svg>
  );
};
