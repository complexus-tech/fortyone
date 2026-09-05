import { cn } from "lib";
import type { Icon } from "./types";

export const ArrowUpRightIcon = ({ className, ...props }: Icon) => (
  <svg
    aria-hidden="true"
    className={cn("h-5 w-auto", className)}
    fill="none"
    height="24"
    stroke="currentColor"
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth="2"
    viewBox="0 0 24 24"
    width="24"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <path d="M7 7h10v10" />
    <path d="M7 17 17 7" />
  </svg>
);
