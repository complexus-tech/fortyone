import { Badge } from "ui";
import { daysFromNow } from "@/lib/format";
import type { WorkspaceSummary } from "@/lib/types";

export const WorkspaceStatusBadge = ({
  workspace,
}: {
  workspace: WorkspaceSummary;
}) => {
  if (workspace.deletedAt) {
    return (
      <Badge color="danger" size="sm" variant="outline">
        Deleted
      </Badge>
    );
  }

  const tier = workspace.subscriptionTier;
  const subscriptionStatus = workspace.subscriptionStatus;
  if (
    tier &&
    tier !== "free" &&
    ["active", "trialing", "past_due"].includes(subscriptionStatus ?? "")
  ) {
    return (
      <Badge color="tertiary" size="sm">
        Paid
      </Badge>
    );
  }

  const days = daysFromNow(workspace.trialEndsOn);
  if (days !== null && days >= 0) {
    return (
      <Badge color="tertiary" size="sm">
        Trial
      </Badge>
    );
  }

  if (days !== null && days < 0) {
    return (
      <Badge color="tertiary" size="sm">
        Expired
      </Badge>
    );
  }

  return (
    <Badge color="tertiary" size="sm">
      Free
    </Badge>
  );
};

export const UserStatusBadge = ({
  isActive,
  isInternal,
}: {
  isActive: boolean;
  isInternal: boolean;
}) => {
  if (!isActive) {
    return (
      <Badge color="danger" size="sm" variant="outline">
        Inactive
      </Badge>
    );
  }

  if (isInternal) {
    return (
      <Badge color="tertiary" size="sm">
        Internal
      </Badge>
    );
  }

  return (
    <Badge color="tertiary" size="sm">
      Active
    </Badge>
  );
};
