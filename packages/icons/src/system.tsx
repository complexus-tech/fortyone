import { cn } from "lib";
import type { Icon } from "./types";

export const SystemIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="currentColor"
      strokeWidth={strokeWidth}
      viewBox="0 0 640 512"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M128 32C92.7 32 64 60.7 64 96v256h64V96h384v256h64V96c0-35.3-28.7-64-64-64H128zM19.2 384c-10.6 0-19.2 8.6-19.2 19.2 0 42.4 34.4 76.8 76.8 76.8h486.4c42.4 0 76.8-34.4 76.8-76.8 0-10.6-8.6-19.2-19.2-19.2H19.2z" />
    </svg>
  );
};
