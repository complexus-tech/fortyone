"use client";

import { useState } from "react";
import { Badge, Box, Button, Flex, Menu, Tooltip } from "ui";
import {
  CommandIcon,
  DocsIcon,
  EmailIcon,
  HelpIcon,
  Notification02Icon,
  PlusIcon,
  ProductTourIcon,
} from "icons";
import { useParams, usePathname } from "next/navigation";
import { cn } from "lib";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { useCurrentAppCommandAction } from "@/components/shared/app-command-action-context";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { WorkspaceActions } from "@/components/shared/workspace-actions";
import { ProfileMenu } from "@/components/shared/sidebar/profile-menu";
import { WorkspacesMenu } from "@/components/shared/sidebar/workspaces-menu";
import { useSidebar } from "@/components/shared/sidebar/sidebar-context";
import { NewObjectiveDialog } from "@/components/ui/new-objective";
import { NewStoryDialog } from "@/components/ui/new-story-dialog";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import {
  useDesktopAction,
  useDesktopDragRegion,
  useDesktopNavigation,
  useDesktopShell,
} from "@/shell/desktop/desktop-provider";
import { DesktopHistoryControls } from "@/shell/desktop/desktop-history-controls";
import { useDesktopHeaderLayout } from "@/shell/desktop/use-desktop-header-layout";
import { Commands } from "./commands";
import { useStoryCreationDialog } from "./use-story-creation-dialog";

const isTeamObjectivesIndex = (pathname: string) =>
  /\/teams\/[^/]+\/objectives\/?$/.test(pathname);

export const AppCommandBar = () => {
  const {
    isOpen: isStoryOpen,
    setIsOpen: setStoryDialogOpen,
    onCreated: handleStoryCreated,
  } = useStoryCreationDialog();
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const { isCollapsed, hasAssistantCards } = useSidebar();
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
  const { startWalkthrough } = useWalkthrough();
  const dragRegion = useDesktopDragRegion();
  const desktopNavigation = useDesktopNavigation();

  const createsObjective =
    pathname === withWorkspace("/roadmap") || isTeamObjectivesIndex(pathname);
  const fallbackLabel = createsObjective
    ? `Create ${getTermDisplay("objectiveTerm")}`
    : `Create ${getTermDisplay("storyTerm")}`;
  const label = action?.label ?? fallbackLabel;
  const isDisabled = action?.disabled ?? userRole === "guest";
  const createsStory = !action && !createsObjective;
  const canCreateStory = userRole === "admin" || userRole === "member";
  const isDesktop = useDesktopShell({
    ready: Boolean(userRole),
    canCreateStory,
    storyLabel: getTermDisplay("storyTerm"),
  });
  const desktopHeaderRef = useDesktopHeaderLayout(isDesktop);
  useDesktopAction("create-story", () => {
    if (canCreateStory) setStoryDialogOpen(true);
  });

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
    setStoryDialogOpen(true);
  };

  return (
    <>
      <div
        className={cn(
          "hidden h-(--app-shell-header-height) shrink-0 items-center md:flex",
          isDesktop ? "desktop-app-header py-0" : "py-3.5",
        )}
        data-app-command-bar
        data-desktop-app-header={isDesktop || undefined}
        ref={desktopHeaderRef}
        {...dragRegion}
      >
        {isDesktop ? (
          <DesktopHistoryControls {...desktopNavigation} />
        ) : (
          <Flex
            align="center"
            className={
              isCollapsed
                ? "w-(--sidebar-width-collapsed) px-2"
                : "w-(--sidebar-width) gap-2 pl-4"
            }
            justify={isCollapsed ? "center" : undefined}
          >
            <WorkspacesMenu isCollapsed={isCollapsed} />
          </Flex>
        )}
        <Flex
          align="center"
          className={cn(
            "min-w-0 flex-1",
            isDesktop
              ? "desktop-app-header-content"
              : "gap-2 pr-4 pl-2 xl:gap-4",
          )}
        >
          <Commands
            className={
              isDesktop
                ? "desktop-app-header-search w-full min-w-0"
                : "max-w-xl flex-1"
            }
            showTrigger
          />
          <Flex
            align="center"
            className={cn(
              "ml-auto shrink-0",
              isDesktop ? "gap-1" : "gap-2 xl:gap-3",
            )}
            data-desktop-header-actions={isDesktop || undefined}
          >
            <WorkspaceActions
              sidebarHasActions={
                !isDesktop && !isCollapsed && !hasAssistantCards
              }
              variant={isDesktop ? "desktop" : "topbar"}
            />
            <Menu>
              <Tooltip title="Help">
                <Menu.Button>
                  <Button
                    aria-label="Help"
                    className={
                      isDesktop
                        ? "h-[32px] gap-1.5 px-2 md:h-[32px]"
                        : "h-11 w-11 justify-center px-0 xl:w-auto xl:px-3"
                    }
                    color="tertiary"
                    data-walkthrough-target={walkthroughTargets.help}
                    leftIcon={
                      <HelpIcon
                        className={isDesktop ? "h-[18px] w-[18px]" : "h-5"}
                      />
                    }
                    variant="naked"
                  >
                    <span
                      className={
                        isDesktop ? undefined : "sr-only xl:not-sr-only"
                      }
                    >
                      Help
                    </span>
                  </Button>
                </Menu.Button>
              </Tooltip>
              <Menu.Items align="end">
                <Menu.Group>
                  <Menu.Item onSelect={startWalkthrough}>
                    <ProductTourIcon />
                    Product tour
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      setIsKeyboardShortcutsOpen(true);
                    }}
                  >
                    <CommandIcon />
                    Keyboard shortcuts
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      window.open(
                        "mailto:info@fortyone.app",
                        "_blank",
                        "noopener,noreferrer",
                      );
                    }}
                  >
                    <EmailIcon />
                    Contact support
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      window.open(
                        "https://docs.fortyone.app",
                        "_blank",
                        "noopener,noreferrer",
                      );
                    }}
                  >
                    <DocsIcon />
                    Documentation
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <Tooltip title="Notifications">
              <Box data-walkthrough-target={walkthroughTargets.notifications}>
                <Button
                  asIcon
                  className={
                    isDesktop
                      ? "group relative h-[32px] w-[32px] md:h-[32px]"
                      : "group relative h-11 w-11"
                  }
                  color="tertiary"
                  href={withWorkspace("/notifications")}
                  leftIcon={
                    <Notification02Icon
                      className={
                        isDesktop
                          ? "h-[18px] w-[18px]"
                          : "h-[1.4rem] transition-transform group-hover:rotate-12"
                      }
                    />
                  }
                  prefetch
                  variant="naked"
                >
                  <span className="sr-only">Notifications</span>
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
              </Box>
            </Tooltip>
            <Button
              aria-label={label}
              className={cn(
                "shrink-0 whitespace-nowrap",
                isDesktop && "ml-1 h-[28px] gap-1 px-2.5 md:h-[28px]",
              )}
              color="primary"
              data-app-contextual-create-button
              data-walkthrough-create-kind={createsStory ? "story" : undefined}
              data-walkthrough-target={walkthroughTargets.create}
              disabled={isDisabled}
              leftIcon={
                <PlusIcon
                  className={cn(
                    "text-current dark:text-current",
                    isDesktop && "h-[18px] w-[18px]",
                  )}
                />
              }
              onClick={handleCreate}
              variant="solid"
            >
              <span className={isDesktop ? "truncate" : undefined}>
                {label}
              </span>
            </Button>
            <ProfileMenu variant={isDesktop ? "desktop" : "topbar"} />
          </Flex>
        </Flex>
      </div>

      <NewStoryDialog
        isOpen={isStoryOpen}
        objectiveId={params.objectiveId}
        onCreated={handleStoryCreated}
        setIsOpen={setStoryDialogOpen}
        sprintId={params.sprintId}
        teamId={params.teamId}
      />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
        teamId={params.teamId}
      />
      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
    </>
  );
};
