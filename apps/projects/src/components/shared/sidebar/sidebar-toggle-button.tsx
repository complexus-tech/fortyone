"use client";

import { ChevronsLeftIcon, ChevronsRightIcon } from "icons";
import { Button, Tooltip } from "ui";
import { useSidebar } from "./sidebar-context";

type SidebarToggleButtonProps = {
  variant?: "default" | "command-bar";
};

export const SidebarToggleButton = ({
  variant = "default",
}: SidebarToggleButtonProps) => {
  const { isCollapsed, toggleSidebar } = useSidebar();
  const label = isCollapsed ? "Expand sidebar" : "Collapse sidebar";
  const isCommandBar = variant === "command-bar";

  return (
    <Tooltip side="bottom" title={label}>
      <Button
        aria-label={label}
        asIcon
        className={
          isCommandBar ? "bg-surface-muted size-8 min-w-8 shrink-0" : "shrink-0"
        }
        color="tertiary"
        onClick={toggleSidebar}
        rounded={isCommandBar ? "full" : undefined}
        size="sm"
        variant="naked"
      >
        {isCollapsed ? (
          <ChevronsRightIcon className="text-foreground" />
        ) : (
          <ChevronsLeftIcon className="text-foreground" />
        )}
      </Button>
    </Tooltip>
  );
};
