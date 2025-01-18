import { cn } from "lib";
import type { Icon } from "./types";

export const ObjectiveIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn(
        "h-[1.15rem] w-auto text-gray dark:text-gray-300",
        className
      )}
      fill="currentColor"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="17.75" cy="6.25" r="4.25" stroke="currentColor" />
      <circle cx="6.25" cy="6.25" r="4.25" stroke="currentColor" />
      <circle cx="17.75" cy="17.75" r="4.25" stroke="currentColor" />
      <circle cx="6.25" cy="17.75" r="4.25" stroke="currentColor" />
    </svg>
  );
};
