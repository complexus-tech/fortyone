"use client";
import { useState } from "react";
import { Badge, Box, Button, Divider, Flex, Tooltip } from "ui";
import { Notification02Icon, PlusIcon, SearchIcon } from "icons";
import { cn } from "lib";
import { useHotkeys } from "react-hotkeys-hook";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import {
  useAnalytics,
  useFeatures,
  useTerminology,
  useWorkspacePath,
} from "@/hooks";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useUserRole } from "@/hooks/role";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { clearAllStorage } from "./utils";
import { WorkspacesMenu } from "./workspaces-menu";
import { logOut } from "./actions";

export const Header = ({ isCollapsed = false }: { isCollapsed?: boolean }) => {
  const { getTermDisplay } = useTerminology();
  const { analytics } = useAnalytics();
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const { userRole } = useUserRole();
  const features = useFeatures();
  const { withWorkspace } = useWorkspacePath();

  useHotkeys("shift+n", () => {
    if (userRole !== "guest") {
      setIsOpen(true);
    }
  });

  useHotkeys("shift+o", () => {
    if (userRole !== "guest" && features.objectiveEnabled) {
      setIsObjectivesOpen(true);
    }
  });

  // Sprint creation shortcut disabled - use team automation settings instead
  // useHotkeys("shift+s", () => {
  //   if (userRole !== "guest" && features.sprintEnabled) {
  //     setIsSprintsOpen(true);
  //   }
  // });

  const handleLogout = async () => {
    const mainDomain =
      process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app"
        ? "https://fortyone.app"
        : "/";
    try {
      await logOut();
      analytics.logout(true);
      clearAllStorage();
      window.location.href = `${mainDomain}?signedOut=true`;
    } catch {
      clearAllStorage();
      window.location.href = `${mainDomain}?signedOut=true`;
    }
  };

  useHotkeys("alt+shift+l", async () => {
    await handleLogout();
  });

  const notificationsAction = (
    <Tooltip side="right" title="Notifications">
      <Box data-sidebar-notifications-button>
        <Button
          asIcon
          className="group relative"
          color="tertiary"
          href={withWorkspace("/notifications")}
          leftIcon={
            <Notification02Icon
              className={cn(
                "transition-transform group-hover:rotate-12",
                "h-6",
              )}
            />
          }
          prefetch
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Notifications</span>
          {unreadNotifications ? (
            <Badge
              className="absolute -top-1 -right-1 shrink-0"
              rounded="full"
              size="sm"
            >
              {unreadNotifications > 9 ? "9+" : unreadNotifications}
            </Badge>
          ) : null}
        </Button>
      </Box>
    </Tooltip>
  );

  return (
    <>
      {isCollapsed ? (
        <>
          <Flex
            align="center"
            className="h-[3.6rem] flex-col pt-2"
            justify="between"
          >
            <WorkspacesMenu isCollapsed />
            <Divider className="w-8 self-center" />
          </Flex>
          <Box className="mt-2 flex justify-center">{notificationsAction}</Box>
        </>
      ) : (
        <Flex align="center" className="h-[3.6rem] pt-2" justify="between">
          <WorkspacesMenu />
          {notificationsAction}
        </Flex>
      )}
      <Flex className={cn("mt-2 mb-3 gap-1.5", isCollapsed && "flex-col")}>
        <Tooltip side="right" title={isCollapsed ? "Create story" : null}>
          <Button
            asIcon={isCollapsed}
            className={cn("truncate md:h-[2.4rem]", isCollapsed && "w-full")}
            color="tertiary"
            data-sidebar-create-story-button
            disabled={userRole === "guest"}
            fullWidth={!isCollapsed}
            leftIcon={<PlusIcon className="shrink-0" />}
            onClick={() => {
              if (userRole !== "guest") {
                setIsOpen(!isOpen);
              }
            }}
            variant="outline"
          >
            <span className={isCollapsed ? "hidden" : undefined}>
              Create {getTermDisplay("storyTerm")}
            </span>
          </Button>
        </Tooltip>
        <Tooltip side="right" title={isCollapsed ? "Search" : null}>
          <Button
            asIcon
            className={cn("md:h-[2.4rem]", isCollapsed && "w-full")}
            color="tertiary"
            href={withWorkspace("/search")}
            leftIcon={<SearchIcon className="h-4" />}
            prefetch
            variant="outline"
          >
            <span className="sr-only">Search</span>
          </Button>
        </Tooltip>
      </Flex>

      {/* Dialogs */}
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} />
      <NewSprintDialog isOpen={isSprintsOpen} setIsOpen={setIsSprintsOpen} />
      <NewObjectiveDialog
        isOpen={isObjectivesOpen}
        setIsOpen={setIsObjectivesOpen}
      />
    </>
  );
};
