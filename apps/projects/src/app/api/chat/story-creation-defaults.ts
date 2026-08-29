import type { WorkspaceCtx } from "@/lib/http";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
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
}: {
  ctx: WorkspaceCtx;
}): Promise<StoryCreationDefaults> => {
  const [preferencesResult, subscriptionResult, workspaceResult] =
    await Promise.allSettled([
      getAutomationPreferences(ctx),
      getSubscription(ctx),
      getWorkspace(ctx),
    ]);
  const autoSchedulingPreferenceEnabled =
    preferencesResult.status === "fulfilled"
      ? preferencesResult.value.autoScheduling
      : false;
  const subscription =
    subscriptionResult.status === "fulfilled" ? subscriptionResult.value : null;
  const trialEndsOn =
    workspaceResult.status === "fulfilled"
      ? workspaceResult.value.trialEndsOn
      : null;
  const hasEligibleSubscription = Boolean(
    subscription &&
      subscription.tier !== "free" &&
      ACTIVE_SUBSCRIPTION_STATUSES.has(subscription.status),
  );
  const autoSchedulingAvailable =
    hasActiveWorkspaceTrial(trialEndsOn) || hasEligibleSubscription;

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
