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
      <Badge color="danger" rounded="full" size="sm" variant="outline">
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
      <Badge color="tertiary" rounded="md" size="sm" variant="outline">
        Paid
      </Badge>
    );
  }

  const days = daysFromNow(workspace.trialEndsOn);
  if (days !== null && days >= 0) {
    return (
      <Badge color="tertiary" rounded="md" size="sm" variant="outline">
        Trial
      </Badge>
    );
  }

  if (days !== null && days < 0) {
    return (
      <Badge color="warning" rounded="full" size="sm" variant="outline">
        Expired
      </Badge>
    );
  }

  return (
    <Badge color="tertiary" rounded="md" size="sm" variant="outline">
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
      <Badge color="danger" rounded="full" size="sm" variant="outline">
        Inactive
      </Badge>
    );
  }

  if (isInternal) {
    return (
      <Badge color="tertiary" rounded="md" size="sm" variant="outline">
        Internal
      </Badge>
    );
  }

  return (
    <Badge color="tertiary" rounded="md" size="sm" variant="outline">
      Active
    </Badge>
  );
};
