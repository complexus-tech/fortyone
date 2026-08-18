"use client";

import { SidebarCollapseIcon, SidebarExpandIcon } from "icons";
import { Button, Tooltip } from "ui";
import { useSidebar } from "./sidebar-context";

export const SidebarToggleButton = () => {
  const { isCollapsed, toggleSidebar } = useSidebar();
  const label = isCollapsed ? "Expand sidebar" : "Collapse sidebar";

  return (
    <Tooltip side="bottom" title={label}>
      <Button
        aria-label={label}
        asIcon
        className="shrink-0"
        color="tertiary"
        onClick={toggleSidebar}
        size="sm"
        variant="naked"
      >
        {isCollapsed ? (
          <SidebarExpandIcon className="text-foreground" />
        ) : (
          <SidebarCollapseIcon className="text-foreground" />
        )}
      </Button>
    </Tooltip>
  );
};
