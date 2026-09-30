"use client";

import { useState } from "react";
import { InviteMembersIcon } from "icons";
import { Button, Flex, Tooltip } from "ui";
import type { SubscriptionTier } from "@/lib/hooks/subscription-features";
import { InviteMembersDialog } from "@/components/ui/invite-members";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

type WorkspaceActionsVariant = "topbar" | "mobile";

type WorkspaceActionsProps = {
  sidebarHasActions?: boolean;
  variant?: WorkspaceActionsVariant;
};

const ACTION_LAYOUTS: Record<WorkspaceActionsVariant, string> = {
  topbar: "shrink-0 items-center gap-2",
  mobile: "w-full flex-col items-stretch gap-2",
};

const getActionVisibility = ({
  canInvite,
  isSubscriptionReady,
  sidebarHasActions,
  tier,
}: {
  canInvite: boolean;
  isSubscriptionReady: boolean;
  sidebarHasActions: boolean;
  tier: SubscriptionTier;
}) => {
  const isLimitedPlan = tier === "free" || tier === "trial";

  return {
    showsInvite: canInvite && (!sidebarHasActions || isLimitedPlan),
    showsSubscriptionStatus:
      isSubscriptionReady && isLimitedPlan && !sidebarHasActions,
  };
};

const InvitePeopleButton = ({
  onInvite,
  variant,
}: {
  onInvite: () => void;
  variant: WorkspaceActionsVariant;
}) => {
  const isMobile = variant === "mobile";

  return (
    <Tooltip title={isMobile ? null : "Invite people"}>
      <Button
        aria-label="Invite people"
        className={
          isMobile
            ? "text-base whitespace-nowrap"
            : "w-[2.1rem] justify-center px-0 text-base whitespace-nowrap lg:w-auto lg:px-2"
        }
        color="tertiary"
        data-invite-button
        fullWidth={isMobile}
        leftIcon={<InviteMembersIcon aria-hidden className="h-5 shrink-0" />}
        onClick={onInvite}
        size="sm"
        variant="outline"
      >
        <span className={isMobile ? undefined : "hidden lg:inline"}>
          Invite people
        </span>
      </Button>
    </Tooltip>
  );
};

const WorkspacePlanStatus = ({
  canManageBilling,
  tier,
  trialDaysRemaining,
  variant,
}: {
  canManageBilling: boolean;
  tier: SubscriptionTier;
  trialDaysRemaining: number;
  variant: WorkspaceActionsVariant;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const isMobile = variant === "mobile";
  const statusLabel =
    tier === "trial"
      ? `${trialDaysRemaining} day${trialDaysRemaining === 1 ? "" : "s"} left in trial`
      : "Free plan";
  const statusActionLabel = tier === "free" ? "Upgrade" : statusLabel;
  const description = canManageBilling
    ? "View plans and manage workspace billing."
    : "Ask your admin to upgrade your workspace.";

  return (
    <Tooltip title={description}>
      <span className={isMobile ? "block w-full" : "block"}>
        {canManageBilling ? (
          <Button
            className="text-text-muted text-base whitespace-nowrap"
            color="tertiary"
            fullWidth={isMobile}
            href={withWorkspace("/settings/workspace/billing")}
            prefetch
            size="sm"
            variant="naked"
          >
            {statusActionLabel}
          </Button>
        ) : (
          <span className="text-text-muted flex h-[2.1rem] items-center px-2 text-base whitespace-nowrap">
            {statusLabel}
          </span>
        )}
      </span>
    </Tooltip>
  );
};

export const WorkspaceActions = ({
  sidebarHasActions = false,
  variant = "topbar",
}: WorkspaceActionsProps) => {
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const { workspace } = useCurrentWorkspace();
  const { userRole } = useUserRole();
  const { isPending, isError } = useSubscription();
  const { tier, trialDaysRemaining } = useSubscriptionFeatures();
  const canInvite = userRole === "admin";
  const { showsInvite, showsSubscriptionStatus } = getActionVisibility({
    canInvite,
    isSubscriptionReady: !isPending && !isError,
    sidebarHasActions,
    tier,
  });
  const showsActions = showsInvite || showsSubscriptionStatus;

  if (!workspace || workspace.deletedAt) return null;
  if (!showsActions && !(canInvite && isInviteOpen)) return null;

  return (
    <>
      {showsActions ? (
        <Flex
          className={ACTION_LAYOUTS[variant]}
          data-workspace-actions={variant}
        >
          {showsInvite ? (
            <InvitePeopleButton
              onInvite={() => {
                setIsInviteOpen(true);
              }}
              variant={variant}
            />
          ) : null}
          {showsSubscriptionStatus ? (
            <WorkspacePlanStatus
              canManageBilling={canInvite}
              tier={tier}
              trialDaysRemaining={trialDaysRemaining}
              variant={variant}
            />
          ) : null}
        </Flex>
      ) : null}
      {canInvite ? (
        <InviteMembersDialog
          isOpen={isInviteOpen}
          setIsOpen={setIsInviteOpen}
        />
      ) : null}
    </>
  );
};
