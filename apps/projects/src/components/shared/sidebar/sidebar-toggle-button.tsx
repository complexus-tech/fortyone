"use client";

import { SidebarCollapseIcon, SidebarExpandIcon } from "icons";
import { cn } from "lib";
import { Kbd, Tooltip } from "ui";
import { useSidebar } from "./sidebar-context";

type SidebarToggleButtonProps = {
  placement?: "command-bar" | "workspace-menu";
};

export const SidebarToggleButton = ({
  placement = "command-bar",
}: SidebarToggleButtonProps) => {
  const { isCollapsed, toggleSidebar } = useSidebar();
  const label = isCollapsed ? "Expand sidebar" : "Collapse sidebar";

  return (
    <Tooltip
      side="bottom"
      title={
        <span className="flex items-center gap-2">
          <span>{label}</span>
          <Kbd className="bg-surface-muted h-5 min-w-5 px-1">⌘ B</Kbd>
        </span>
      }
    >
      <button
        aria-label={label}
        className={cn(
          "text-foreground focus-visible:ring-ring flex shrink-0 cursor-pointer items-center justify-center border-0 bg-transparent p-0 transition-colors focus-visible:ring-2 focus-visible:outline-none",
          placement === "workspace-menu" ? "rounded-xl" : "rounded-full",
        )}
        data-sidebar-toggle-location={placement}
        onClick={toggleSidebar}
        style={{
          flexBasis: "38px",
          height: "38px",
          maxWidth: "38px",
          minWidth: "38px",
          width: "38px",
        }}
        type="button"
      >
        {isCollapsed ? (
          <SidebarExpandIcon className="h-5.5" />
        ) : (
          <SidebarCollapseIcon className="h-5.5" />
        )}
      </button>
    </Tooltip>
  );
};
