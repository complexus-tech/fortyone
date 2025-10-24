"use client";
import { Box, Button, Flex, Menu, Text, Tooltip } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { addHours, differenceInHours } from "date-fns";
import { InviteMembersDialog } from "@/components/ui";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole } from "@/hooks";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { Commands } from "../commands";
import { CommandBar } from "../command-bar";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";
import { ProfileMenu } from "./profile-menu";

export const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const [isCommandBarOpen, setIsCommandBarOpen] = useState(false);
  const { workspace } = useCurrentWorkspace();

  const { tier, trialDaysRemaining } = useSubscriptionFeatures();
  const { userRole } = useUserRole();

  const getTimeRemaining = () => {
    if (!workspace?.deletedAt) return null;
    const hoursRemaining = differenceInHours(
      addHours(new Date(workspace.deletedAt), 48),
      new Date(),
    );
    if (hoursRemaining <= 0) return null;
    return hoursRemaining;
  };

  return (
    <Box className="relative flex h-dvh flex-col justify-between bg-gradient-to-br from-gray-100/50 to-gray-50/50 pb-6 dark:bg-gradient-to-tl dark:from-[#000000] dark:to-dark-200">
      <Box className="relative z-[1] px-4">
        <Header />
        <Navigation />
        <Teams />
      </Box>
      <Box className="relative z-[1]">
        <Box className="mb-2.5 px-3.5">
          {workspace?.deletedAt ? (
            <Box className="mb-4 rounded-xl border-[0.5px] border-warning bg-warning/20 p-4 shadow-lg shadow-gray-100 dark:border-warning/20 dark:bg-warning/10 dark:shadow-none">
              <Text className="dark:text-white" fontWeight="semibold">
                Workspace scheduled for deletion
              </Text>
              {getTimeRemaining() ? (
                <Text className="mt-1 opacity-80">
                  Your workspace is scheduled for deletion in about{" "}
                  {getTimeRemaining()} hour{getTimeRemaining() !== 1 ? "s" : ""}
                  .
                </Text>
              ) : (
                <Text className="mt-1 opacity-80">
                  Your workspace has been scheduled for deletion and may be
                  deleted at any time.
                </Text>
              )}
              {userRole === "admin" && (
                <Button
                  className="mt-3 px-3"
                  color="warning"
                  href="/settings"
                  prefetch
                  size="sm"
                >
                  Restore workspace
                </Button>
              )}
            </Box>
          ) : (
            <>
              {tier === "free" && (
                <Box className="rounded-xl border-[0.5px] border-gray-200/60 bg-white p-4 shadow-lg shadow-gray-100 dark:border-dark-50 dark:bg-dark-300 dark:shadow-none">
                  <Text fontWeight="medium">You&apos;re on the free plan</Text>
                  <Text className="mt-2" color="muted">
                    {userRole === "admin"
                      ? "Upgrade to a paid plan to get more features."
                      : "Ask your admin to upgrade to a paid plan to get more features."}
                  </Text>
                  {userRole === "admin" && (
                    <Button
                      className="mt-3 px-3"
                      color="tertiary"
                      href="/settings/workspace/billing"
                      prefetch
                      size="sm"
                    >
                      Upgrade plan
                    </Button>
                  )}
                </Box>
              )}
              {tier === "trial" && (
                <Tooltip
                  className="ml-2 max-w-56 py-3"
                  title={`${trialDaysRemaining} days left in your trial. ${userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade"} to a paid plan to get more premium features.`}
                >
                  <span>
                    <Button
                      className="mt-3 border-opacity-15 bg-opacity-10 px-3 text-primary dark:bg-opacity-15"
                      href={
                        userRole === "admin"
                          ? "/settings/workspace/billing"
                          : undefined
                      }
                      prefetch
                      rounded="lg"
                      size="sm"
                    >
                      {trialDaysRemaining} day
                      {trialDaysRemaining !== 1 ? "s" : ""} left in trial
                    </Button>
                  </span>
                </Tooltip>
              )}
            </>
          )}

          {!workspace?.deletedAt ? (
            <Flex align="center" className="mt-3" justify="between">
              {userRole === "admin" ? (
                <button
                  className="flex items-center justify-start gap-2 px-1 text-left"
                  data-invite-button
                  onClick={() => {
                    setIsOpen(true);
                  }}
                  type="button"
                >
                  <PlusIcon />
                  <span className="line-clamp-1">Invite members</span>
                </button>
              ) : (
                <Button
                  color="tertiary"
                  onClick={() => {
                    setIsCommandBarOpen(true);
                  }}
                  size="xs"
                >
                  <CommandIcon className="h-4" /> K
                </Button>
              )}

              <Menu>
                <Menu.Button>
                  <Button
                    asIcon
                    className="border-[0.5px]"
                    color="tertiary"
                    data-help-button
                    rounded="full"
                    variant="naked"
                  >
                    <HelpIcon className="h-6" />
                  </Button>
                </Menu.Button>
                <Menu.Items align="end">
                  <Menu.Group>
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
                        window.open("mailto:support@complexus.app", "_blank");
                      }}
                    >
                      <EmailIcon />
                      Contact support
                    </Menu.Item>
                    <Menu.Item
                      onSelect={() => {
                        window.open("https://docs.fortyone.app", "_blank");
                      }}
                    >
                      <DocsIcon />
                      Documentation
                    </Menu.Item>
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </Flex>
          ) : null}
        </Box>
        <ProfileMenu />
      </Box>

      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
      <InviteMembersDialog isOpen={isOpen} setIsOpen={setIsOpen} />
      <Commands />
      <CommandBar isOpen={isCommandBarOpen} setIsOpen={setIsCommandBarOpen} />
    </Box>
  );
};
