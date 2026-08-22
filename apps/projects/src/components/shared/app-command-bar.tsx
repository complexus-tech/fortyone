"use client";

import { useState } from "react";
import { Badge, Button, Flex, Tooltip } from "ui";
import { Notification02Icon, PlusIcon } from "icons";
import { useParams, usePathname } from "next/navigation";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { useTerminology, useUserRole, useWorkspacePath } from "@/hooks";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { Commands } from "@/components/shared/commands";
import { ProfileMenu } from "@/components/shared/sidebar/profile-menu";
import { SidebarToggleButton } from "@/components/shared/sidebar/sidebar-toggle-button";
import { WorkspacesMenu } from "@/components/shared/sidebar/workspaces-menu";
import { useSidebar } from "@/components/shared/sidebar/sidebar-context";
import { useCurrentAppCommandAction } from "./app-command-action-context";

const isTeamObjectivesIndex = (pathname: string) =>
  /\/teams\/[^/]+\/objectives\/?$/.test(pathname);

export const AppCommandBar = () => {
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const { isCollapsed } = useSidebar();
  const action = useCurrentAppCommandAction();
  const pathname = usePathname();
  const params = useParams<{
    objectiveId?: string;
    sprintId?: string;
    teamId?: string;
  }>();
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();

  const createsObjective =
    pathname === withWorkspace("/roadmap") || isTeamObjectivesIndex(pathname);
  const fallbackLabel = createsObjective
    ? `Create ${getTermDisplay("objectiveTerm")}`
    : `Create ${getTermDisplay("storyTerm")}`;
  const label = action?.label ?? fallbackLabel;
  const isDisabled = action?.disabled ?? userRole === "guest";

  const handleCreate = () => {
    if (isDisabled) return;
    if (action) {
      action.onSelect();
      return;
    }
    if (createsObjective) {
      setIsObjectiveOpen(true);
      return;
    }
    setIsStoryOpen(true);
  };

  return (
    <>
      <Flex
        align="center"
        className="hidden h-[60px] shrink-0 py-[10px] md:flex"
        data-app-command-bar
      >
        <Flex
          align="center"
          className={
            isCollapsed
              ? "w-(--sidebar-width-collapsed) px-2"
              : "w-(--sidebar-width) px-4"
          }
          justify={isCollapsed ? "center" : undefined}
        >
          <WorkspacesMenu isCollapsed={isCollapsed} />
        </Flex>
        <Flex align="center" className="min-w-0 flex-1 gap-4 pr-[16px] pl-2">
          <Flex align="center" className="h-11 w-11 shrink-0" justify="center">
            <SidebarToggleButton variant="command-bar" />
          </Flex>
          <Commands className="max-w-xl flex-1" showTrigger />
          <Flex align="center" className="ml-auto shrink-0 gap-4">
            <Tooltip title="Notifications">
              <Button
                aria-label="Notifications"
                asIcon
                className="group relative h-11 w-11"
                color="tertiary"
                href={withWorkspace("/notifications")}
                leftIcon={
                  <Notification02Icon className="h-[1.4rem] transition-transform group-hover:rotate-12" />
                }
                prefetch
                variant="naked"
              >
                {unreadNotifications ? (
                  <Badge
                    className="absolute -top-0.5 -right-0.5 shrink-0"
                    rounded="full"
                    size="sm"
                  >
                    {unreadNotifications > 9 ? "9+" : unreadNotifications}
                  </Badge>
                ) : null}
              </Button>
            </Tooltip>
            <Tooltip title={label}>
              <Button
                aria-label={label}
                asIcon
                className="size-11 max-w-11 min-w-11 shrink-0"
                color="primary"
                data-app-contextual-create-button
                disabled={isDisabled}
                leftIcon={
                  <PlusIcon className="text-current dark:text-current" />
                }
                onClick={handleCreate}
                rounded="full"
                style={{
                  flexBasis: "32px",
                  height: "32px",
                  maxWidth: "32px",
                  minWidth: "32px",
                  width: "32px",
                }}
                variant="solid"
              />
            </Tooltip>
            <ProfileMenu variant="topbar" />
          </Flex>
        </Flex>
      </Flex>

      <NewStoryDialog
        isOpen={isStoryOpen}
        objectiveId={params.objectiveId}
        setIsOpen={setIsStoryOpen}
        sprintId={params.sprintId}
        teamId={params.teamId}
      />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
        teamId={params.teamId}
      />
    </>
  );
};
