import { cn } from "lib";
import type { Icon } from "./types";

export const ItalicIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M12 4H19" stroke="currentColor" strokeLinecap="round" />
      <path d="M8 20L16 4" stroke="currentColor" strokeLinecap="round" />
      <path d="M5 20H12" stroke="currentColor" strokeLinecap="round" />
    </svg>
  );
};
