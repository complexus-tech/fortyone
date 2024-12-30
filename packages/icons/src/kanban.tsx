import { cn } from "lib";
import type { Icon } from "./types";

export const KanbanIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
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
      <path d="M10 21.75H14V2.25H10V21.75Z" fill="currentColor" />
      <path
        d="M8 2.25H3C2.58579 2.25 2.25 2.58579 2.25 3V21C2.25 21.4142 2.58579 21.75 3 21.75H8V2.25Z"
        fill="currentColor"
      />
      <path
        d="M16 2.25V21.75H21C21.4142 21.75 21.75 21.4142 21.75 21V3C21.75 2.58579 21.4142 2.25 21 2.25H16Z"
        fill="currentColor"
      />
    </svg>
  );
};
