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
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M12 8V12" />
      <path d="M12 16h.01" />
    </svg>
  );
};
