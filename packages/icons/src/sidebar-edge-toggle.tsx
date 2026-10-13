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
    <g className="sidebar-edge-toggle-shape" data-direction={direction}>
      <path
        className="sidebar-edge-toggle-segment"
        d="M12 4V12"
        data-segment="top"
      />
      <path
        className="sidebar-edge-toggle-segment"
        d="M12 12V20"
        data-segment="bottom"
      />
    </g>
  </svg>
);
