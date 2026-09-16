import type { WorkspaceCtx } from "@/lib/http";
import type { Subscription, Workspace } from "@/types";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { DEFAULT_TIME_NEEDED_MINUTES } from "@/lib/time-needed";

export type StoryCreationDefaults = {
  autoSchedulingAvailable: boolean;
  bulkStories: {
    autoSchedulingEnabled: false;
    estimatedDurationMinutes: null;
  };
  singleStory: {
    autoSchedulingEnabled: boolean;
    estimatedDurationMinutes: number;
  };
};

const ACTIVE_SUBSCRIPTION_STATUSES = new Set([
  "active",
  "trialing",
  "past_due",
]);

const hasActiveWorkspaceTrial = (trialEndsOn: string | null | undefined) => {
  if (!trialEndsOn) return false;

  const trialEnd = new Date(trialEndsOn);
  return !Number.isNaN(trialEnd.getTime()) && trialEnd > new Date();
};

export const resolveStoryCreationDefaults = async ({
  ctx,
  workspace,
  subscription,
}: {
  ctx: WorkspaceCtx;
  workspace: Pick<Workspace, "trialEndsOn">;
  subscription: Subscription | null;
}): Promise<StoryCreationDefaults> => {
  // Reuse the context already authorized and loaded for this request.
  const [preferencesResult] = await Promise.allSettled([
    getAutomationPreferences(ctx),
  ]);
  const autoSchedulingPreferenceEnabled =
    preferencesResult.status === "fulfilled"
      ? preferencesResult.value.autoScheduling
      : false;
  const hasEligibleSubscription = Boolean(
    subscription &&
      subscription.tier !== "free" &&
      ACTIVE_SUBSCRIPTION_STATUSES.has(subscription.status),
  );
  const autoSchedulingAvailable =
    hasActiveWorkspaceTrial(workspace.trialEndsOn) || hasEligibleSubscription;

  return {
    autoSchedulingAvailable,
    bulkStories: {
      autoSchedulingEnabled: false,
      estimatedDurationMinutes: null,
    },
    singleStory: {
      autoSchedulingEnabled:
        autoSchedulingAvailable && autoSchedulingPreferenceEnabled,
      estimatedDurationMinutes: DEFAULT_TIME_NEEDED_MINUTES,
    },
  };
};
