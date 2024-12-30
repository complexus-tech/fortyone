import { cn } from "lib";
import type { Icon } from "./types";

export const StoryIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="currentColor"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M15.25 5c0 -2.07106 1.6789 -3.75 3.75 -3.75S22.75 2.92894 22.75 5 21.0711 8.75 19 8.75 15.25 7.07106 15.25 5Z"
        fill="currentColor"
        strokeWidth={1}
      />
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M14.527 2.25c-0.4928 0.79977 -0.777 1.74169 -0.777 2.75 0 2.89947 2.3504 5.25 5.25 5.25 1.0083 0 1.9502 -0.28425 2.75 -0.77699V15c0 3.7279 -3.0221 6.75 -6.75 6.75H9c-3.72792 0 -6.75 -3.0221 -6.75 -6.75V9c0 -3.72792 3.02208 -6.75 6.75 -6.75h5.527Z"
        fill="currentColor"
        strokeWidth={1}
      />
    </svg>
  );
};
