import { cn } from "lib";
import type { Icon } from "./types";

type SidebarEdgeToggleIconProps = Icon & {
  direction: "left" | "right";
};

export const SidebarEdgeToggleIcon = ({
  className,
  direction,
  strokeWidth = 2,
  ...rest
}: SidebarEdgeToggleIconProps) => (
  <svg
    {...rest}
    className={cn("h-8 w-auto text-icon", className)}
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
    <path
      className="sidebar-edge-toggle-path"
      d="M12 4L12 12L12 20"
      data-direction={direction}
    />
  </svg>
);
