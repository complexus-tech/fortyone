"use client";
import type { ReactNode } from "react";
import { Box, Button, Flex, Menu, Text, Tooltip } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon } from "icons";
import { useState } from "react";
import { addHours, differenceInHours } from "date-fns";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { Commands } from "../commands";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";
import { ProfileMenu } from "./profile-menu";
import { UpcomingMeetingCard } from "./upcoming-meeting-card";

const SidebarFooterActions = ({
  leadingAction,
  onOpenKeyboardShortcuts,
}: {
  leadingAction: ReactNode;
  onOpenKeyboardShortcuts: () => void;
}) => (
  <Flex align="center" className="mt-3">
    <Box className="min-w-0 flex-1">{leadingAction}</Box>
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
          <Menu.Item onSelect={onOpenKeyboardShortcuts}>
            <CommandIcon />
            Keyboard shortcuts
          </Menu.Item>
          <Menu.Item
            onSelect={() => {
              window.open(
                "mailto:support@complexus.app",
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
  </Flex>
);

export const Sidebar = () => {
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const { workspace } = useCurrentWorkspace();
  const { withWorkspace } = useWorkspacePath();

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
  const upgradeAction =
    userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade";
  const subscriptionTitle =
    tier === "trial"
      ? `${trialDaysRemaining} days left in your trial. ${upgradeAction} to a paid plan to get more premium features.`
      : `You are on the free plan. ${upgradeAction} to a paid plan to get more features.`;
  const subscriptionLabel =
    tier === "trial"
      ? `${trialDaysRemaining} day${trialDaysRemaining !== 1 ? "s" : ""} left in trial`
      : "Upgrade";
  const subscriptionAction =
    tier === "free" || tier === "trial" ? (
      <Tooltip className="ml-2 max-w-56 py-3" title={subscriptionTitle}>
        <span>
          <Button
            className="text-primary border-primary/15 bg-primary/15 dark:bg-primary/10 dark:bg-border-primary/15 px-2.5"
            href={
              userRole === "admin"
                ? withWorkspace("/settings/workspace/billing")
                : undefined
            }
            prefetch
            rounded="lg"
            size="sm"
          >
            {subscriptionLabel}
          </Button>
        </span>
      </Tooltip>
    ) : null;
  const openKeyboardShortcuts = () => {
    setIsKeyboardShortcutsOpen(true);
  };

  return (
    <Box className="bg-sidebar border-border/80 dark:bg-sidebar/40 relative flex h-dvh w-(--sidebar-width) shrink-0 flex-col overflow-hidden border-r-[0.5px]">
      <Box className="relative z-1 shrink-0 px-4" data-sidebar-header>
        <Header />
      </Box>
      <Box
        className="relative z-1 min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pb-3"
        data-sidebar-content
      >
        <Navigation />
        <Teams />
      </Box>
      <Box className="relative z-1 shrink-0 pb-4" data-sidebar-footer>
        <Box className="mb-2.5 px-3.5">
          {workspace?.deletedAt ? (
            <Box className="border-warning bg-warning/20 shadow-shadow mb-4 rounded-xl border-[0.5px] p-4 shadow-lg">
              <Text className="text-foreground" fontWeight="semibold">
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
                  href={withWorkspace("/settings")}
                  prefetch
                  size="sm"
                >
                  Restore workspace
                </Button>
              )}
            </Box>
          ) : (
            <UpcomingMeetingCard
              fallback={
                <SidebarFooterActions
                  leadingAction={subscriptionAction}
                  onOpenKeyboardShortcuts={openKeyboardShortcuts}
                />
              }
            />
          )}
        </Box>
        <ProfileMenu />
      </Box>

      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
      <Commands />
    </Box>
  );
};
