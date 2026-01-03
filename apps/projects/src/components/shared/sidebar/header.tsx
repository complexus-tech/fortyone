"use client";
import { useState } from "react";
import { Button, Flex, Badge, Box, Menu } from "ui";
import {
  NewStoryIcon,
  SearchIcon,
  BellIcon,
  PlusIcon,
  ArrowDown2Icon,
} from "icons";
import { useHotkeys } from "react-hotkeys-hook";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { useAnalytics, useFeatures, useTerminology } from "@/hooks";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useUserRole } from "@/hooks/role";
import { useUnreadNotifications } from "@/modules/notifications/hooks/unread";
import { clearAllStorage } from "./utils";
import { WorkspacesMenu } from "./workspaces-menu";
import { logOut } from "./actions";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

export const Header = () => {
  const { getTermDisplay } = useTerminology();
  const { analytics } = useAnalytics();
  const [isOpen, setIsOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const { userRole } = useUserRole();
  const features = useFeatures();

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

  useHotkeys("alt+shift+l", async () => {
    await handleLogout();
  });

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
      clearAllStorage();
      window.location.href = `https://www.${domain}?signedOut=true`;
    } finally {
      clearAllStorage();
      window.location.href = `https://www.${domain}?signedOut=true`;
    }
  };

  return (
    <>
      <Flex align="center" className="h-16" justify="between">
        <WorkspacesMenu />
        <Box data-sidebar-notifications-button>
          <Button
            asIcon
            className="group relative"
            color="tertiary"
            href="/notifications"
            leftIcon={
              <BellIcon className="h-[1.4rem] transition-transform group-hover:rotate-12" />
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
      </Flex>
      <Flex className="mb-3 gap-1">
        <Button
          className="truncate md:h-[2.4rem]"
          color="tertiary"
          data-sidebar-create-story-button
          disabled={userRole === "guest"}
          fullWidth
          leftIcon={<PlusIcon className="shrink-0" />}
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(!isOpen);
            }
          }}
          variant="outline"
        >
          Create {getTermDisplay("storyTerm")}
        </Button>
        <Button
          asIcon
          className="md:h-[2.4rem]"
          color="tertiary"
          href="/search"
          leftIcon={<SearchIcon className="h-4" />}
          prefetch
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
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
