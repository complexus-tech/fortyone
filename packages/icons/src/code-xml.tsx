import { cn } from "lib";
import type { Icon } from "./types";

export const CodeXmlIcon = (props: Icon) => {
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
      <path d="M15 4L9 20" />
      <path d="M5.99997 16C5.99997 16 2.00001 13.054 2 12C1.99999 10.9459 6 8 6 8" />
      <path d="M18 8C18 8 22 10.946 22 12C22 13.0541 18 16 18 16" />
    </svg>
  );
};
