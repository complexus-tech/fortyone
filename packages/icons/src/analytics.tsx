import type { Icon } from "./types";
import { cn } from "lib";

export const AnalyticsIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
  return (
    <svg
      className={cn(
        "shrink-0 text-icon h-5 w-auto",
        className
      )}
      fill="currentColor"
      focusable="false"
      strokeWidth={strokeWidth}
      {...rest}
      height="16"
      viewBox="0 0 16 16"
      width="16"
    >
      <rect height="6" rx="1" width="3" x="1" y="8" />
      <rect height="9" rx="1" width="3" x="6" y="5" />
      <rect height="12" rx="1" width="3" x="11" y="2" />
    </svg>
  );
};
