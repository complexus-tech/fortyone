import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Notification02Icon, Stroke Rounded (MIT).
export const Notification02Icon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M19 18V9.5C19 5.63401 15.866 2.5 12 2.5C8.13401 2.5 5 5.63401 5 9.5V18"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M20.5 18H3.5" strokeLinecap="round" strokeLinejoin="round" />
      <path
        d="M13.5 20C13.5 20.8284 12.8284 21.5 12 21.5M10.5 20C10.5 20.8284 11.1716 21.5 12 21.5M12 21.5V20"
        strokeLinejoin="round"
      />
    </svg>
  );
};
