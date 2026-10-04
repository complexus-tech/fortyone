"use client";
import { Box, Button, Text, Tooltip } from "ui";
import { PlusIcon } from "icons";
import { cn } from "lib";
import { useState } from "react";
import { addHours, differenceInHours } from "date-fns";
import { InviteMembersDialog } from "@/components/ui";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { ViewFavoritesSlot } from "@/shared/views/favorites-slot";
import { Navigation } from "./navigation";
import { Teams } from "./teams";
import { SidebarAssistantCards } from "./upcoming-meeting-card";
import { useSidebar } from "./sidebar-context";

const SidebarSubscriptionAction = ({
  href,
  isCollapsed,
  isOnTrial,
  trialDaysRemaining,
}: {
  href?: string;
  isCollapsed: boolean;
  isOnTrial: boolean;
  trialDaysRemaining: number;
}) => {
  const upgradeAction = href ? "Upgrade" : "Ask your admin to upgrade";
  const isTrialCollapsed = isOnTrial && isCollapsed;
  const trialDaysLabel = `${trialDaysRemaining} day${trialDaysRemaining !== 1 ? "s" : ""}`;
  const fullTrialLabel = `${trialDaysLabel} left in trial`;
  const subscriptionTitle = isOnTrial
    ? `${trialDaysLabel} left in your trial. ${upgradeAction} to a paid plan to get more premium features.`
    : `You are on the free plan. ${upgradeAction} to a paid plan to get more features.`;
  const trialLabel = isCollapsed
    ? `${trialDaysRemaining}d left`
    : fullTrialLabel;
  const subscriptionLabel = isOnTrial ? trialLabel : "Upgrade";
  return (
    <Tooltip className="ml-2 max-w-56 py-3" title={subscriptionTitle}>
      <span className={isOnTrial ? "block max-w-full min-w-0" : undefined}>
        <Button
          align={isTrialCollapsed ? "center" : undefined}
          className={cn(
            "text-primary border-primary/15 bg-primary/15 dark:bg-primary/10 dark:bg-border-primary/15 px-2.5",
            isOnTrial && "max-w-full min-w-0 truncate",
            isTrialCollapsed && "px-1 text-center",
          )}
          fullWidth={isTrialCollapsed}
          href={href}
          prefetch
          rounded="lg"
          size="sm"
        >
          {isOnTrial ? (
            <>
              <span
                aria-hidden={isCollapsed || undefined}
                className="min-w-0 truncate"
              >
                {subscriptionLabel}
              </span>
              {isCollapsed ? (
                <span className="sr-only">{fullTrialLabel}</span>
              ) : null}
            </>
          ) : (
            subscriptionLabel
          )}
        </Button>
      </span>
    </Tooltip>
  );
};

export const Sidebar = () => {
  const [isInviteMembersOpen, setIsInviteMembersOpen] = useState(false);
  const { workspace } = useCurrentWorkspace();
  const { withWorkspace } = useWorkspacePath();
  const { isCollapsed } = useSidebar();

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
  const isOnTrial = tier === "trial";
  const showsUpgradeAction = tier === "free" || isOnTrial;
  const subscriptionAction = showsUpgradeAction ? (
    <SidebarSubscriptionAction
      href={
        userRole === "admin"
          ? withWorkspace("/settings/workspace/billing")
          : undefined
      }
      isCollapsed={isCollapsed}
      isOnTrial={isOnTrial}
      trialDaysRemaining={trialDaysRemaining}
    />
  ) : null;
  const inviteMembersAction =
    userRole === "admin" && !showsUpgradeAction ? (
      <button
        className="flex items-center justify-start gap-2 px-1 text-left"
        data-invite-button
        onClick={() => {
          setIsInviteMembersOpen(true);
        }}
        type="button"
      >
        <PlusIcon />
        <span className="line-clamp-1">Invite members</span>
      </button>
    ) : null;
  const sidebarAction =
    tier === "free" ? subscriptionAction : inviteMembersAction;
  return (
    <Box
      className={cn(
        "relative flex h-full w-(--sidebar-width) shrink-0 flex-col overflow-hidden transition-[width] duration-200 ease-linear",
        isCollapsed && "w-(--sidebar-width-collapsed)",
      )}
      data-sidebar-collapsed={isCollapsed ? "true" : "false"}
    >
      <Box
        className={cn(
          "relative z-1 min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pt-2 pb-3",
          isCollapsed && "px-3",
        )}
        data-sidebar-content
      >
        <Navigation isCollapsed={isCollapsed} />
        <ViewFavoritesSlot isCollapsed={isCollapsed} />
        <Teams isCollapsed={isCollapsed} />
      </Box>
      <Box
        className={cn("relative z-1 shrink-0 pb-4", isCollapsed && "px-2")}
        data-sidebar-footer
      >
        <Box className={cn("mb-2.5 px-3.5", isCollapsed && "px-0")}>
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
            <SidebarAssistantCards
              fallback={
                !isCollapsed && sidebarAction ? (
                  <Box className="mt-3">{sidebarAction}</Box>
                ) : null
              }
              isCollapsed={isCollapsed}
            />
          )}
        </Box>
        {isOnTrial && !workspace?.deletedAt ? (
          <Box
            className={cn("px-3.5", isCollapsed && "px-0")}
            data-sidebar-trial
          >
            {subscriptionAction}
          </Box>
        ) : null}
      </Box>
      <InviteMembersDialog
        isOpen={isInviteMembersOpen}
        setIsOpen={setIsInviteMembersOpen}
      />
    </Box>
  );
};
