"use client";

import { SidebarEdgeToggleIcon } from "icons";
import { Tooltip } from "ui";
import { useSidebar } from "./sidebar-context";

export const SidebarEdgeToggle = () => {
  const { isCollapsed, toggleSidebar } = useSidebar();
  const label = isCollapsed ? "Expand sidebar" : "Collapse sidebar";
  const tooltipLabel = isCollapsed ? "Expand" : "Collapse";

  return (
    <Tooltip
      delayDuration={250}
      side="right"
      sideOffset={2}
      title={tooltipLabel}
    >
      <button
        aria-label={label}
        className="group focus-visible:ring-ring absolute top-[42%] right-0 z-20 flex size-10 translate-x-1/2 -translate-y-1/2 cursor-pointer items-center justify-center rounded-lg border-0 bg-transparent focus-visible:ring-2 focus-visible:outline-none"
        data-sidebar-toggle-location="edge"
        onClick={toggleSidebar}
        type="button"
      >
        <SidebarEdgeToggleIcon
          aria-hidden="true"
          className="text-foreground/70 group-hover:text-foreground group-focus-visible:text-foreground transition-colors motion-reduce:transition-none"
          direction={isCollapsed ? "right" : "left"}
          strokeWidth={2}
        />
      </button>
    </Tooltip>
  );
};
