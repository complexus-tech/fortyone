import { differenceInDays, isAfter } from "date-fns";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

/**
 * Feature limits for each subscription tier
 * Each tier defines maximum limits and feature availability
 */
export const TIER_LIMITS = {
  trial: {
    maxMembers: 50,
    maxFileUploads: "Unlimited",
    maxStories: Infinity,
    maxTeams: 3,
    customTerminology: true,
    privateTeams: true,
    customWorkflows: true,
    maxObjectives: 3,
    objective: true,
  },
  free: {
    maxMembers: 5,
    maxFileUploads: "10MB",
    maxStories: 50,
    maxTeams: 1,
    customTerminology: false,
    privateTeams: false,
    customWorkflows: false,
    maxObjectives: 0,
    objective: false,
  },
  pro: {
    maxMembers: 50,
    maxFileUploads: "Unlimited",
    maxStories: Infinity,
    maxTeams: 3,
    customTerminology: false,
    privateTeams: false,
    customWorkflows: true,
    maxObjectives: 3,
    objective: true,
  },
  business: {
    maxMembers: Infinity,
    maxFileUploads: "Unlimited",
    maxStories: Infinity,
    maxTeams: Infinity,
    customTerminology: true,
    privateTeams: true,
    customWorkflows: true,
    maxObjectives: Infinity,
    objective: true,
  },
  enterprise: {
    maxMembers: Infinity,
    maxFileUploads: "Unlimited",
    maxStories: Infinity,
    maxTeams: Infinity,
    customTerminology: true,
    privateTeams: true,
    customWorkflows: true,
    maxObjectives: Infinity,
    objective: true,
  },
} as const;

export type SubscriptionTier = keyof typeof TIER_LIMITS;

const ACTIVE_SUBSCRIPTION_STATUSES = ["active", "trialing", "past_due"];

/**
 * Hook for managing subscription-based feature access
 *
 * Determines the current subscription tier and provides utility functions
 * to check if specific features are available or if limits are exceeded.
 * Takes into account trial periods and different subscription levels.
 *
 * @returns An object containing feature access utilities and subscription information:
 * - tier: The effective subscription tier (including trial status)
 * - displayTier: Human-readable tier name (includes trial indicator)
 * - isOnTrial: Whether the user is currently on a trial period
 * - trialDaysRemaining: Number of days left in trial period
 * - trialEndsOn: Trial end date ISO string if applicable
 * - hasFeature: Check if a feature is available
 * - getLimit: Get the limit value for a feature
 * - withinLimit: Check if a count is within feature limits
 * - remaining: Calculate remaining capacity for a feature
 * - isTierAtLeast: Check if current tier meets minimum level
 */
export const useSubscriptionFeatures = () => {
  const { data: subscription } = useSubscription();
  const { workspace } = useCurrentWorkspace();

  const isOnTrial =
    workspace?.trialEndsOn &&
    isAfter(new Date(workspace.trialEndsOn), new Date());

  let effectiveTier: SubscriptionTier = "free";

  if (isOnTrial) {
    effectiveTier = "trial";
  }

  if (
    subscription &&
    ACTIVE_SUBSCRIPTION_STATUSES.includes(subscription.status)
  ) {
    effectiveTier = subscription.tier;
  }

  // Get the limits for the current tier
  const limits = TIER_LIMITS[effectiveTier];

  const trialDaysRemaining =
    (workspace?.trialEndsOn
      ? Math.max(
          0,
          differenceInDays(new Date(workspace.trialEndsOn), new Date()),
        )
      : 0) + 1;

  const subscriptionTierName =
    subscription?.tier === "free" ? "hobby" : subscription?.tier || "hobby";

  return {
    tier: effectiveTier,
    billingInterval: subscription?.billingInterval,
    billingEndsAt: subscription?.billingEndsAt,
    displayTier: subscriptionTierName,
    isOnTrial,
    trialDaysRemaining,
    trialEndsOn: workspace?.trialEndsOn,

    /**
     * Check if a specific feature is available in the current subscription tier
     * @param feature - The feature to check
     * @returns Whether the feature is available
     */
    hasFeature: (feature: keyof typeof TIER_LIMITS.free): boolean => {
      if (typeof limits[feature] === "boolean") {
        return limits[feature] as boolean;
      }
      return (limits[feature] as number) > 0;
    },

    /**
     * Get the limit value for a specific feature
     * @param feature - The feature to get the limit for
     * @returns The limit value (could be number, string, or boolean)
     */
    getLimit: <T extends keyof typeof TIER_LIMITS.free>(feature: T) => {
      return limits[feature];
    },

    /**
     * Check if a count is within the limit for a feature
     * @param feature - The feature to check
     * @param count - The current count to check against the limit
     * @returns Whether the count is within limits
     */
    withinLimit: (
      feature: keyof typeof TIER_LIMITS.free,
      count: number,
    ): boolean => {
      const limit = limits[feature];
      return typeof limit === "number"
        ? count < limit || limit === Infinity
        : true;
    },

    /**
     * Calculate how many more items can be added before reaching the limit
     * @param feature - The feature to check
     * @param currentCount - The current usage count
     * @returns Remaining capacity (0 if already at/over limit)
     */
    remaining: (
      feature: keyof typeof TIER_LIMITS.free,
      currentCount: number,
    ): number => {
      const limit = limits[feature];
      if (typeof limit !== "number") return 0;
      return limit === Infinity ? Infinity : Math.max(0, limit - currentCount);
    },

    /**
     * Check if the current subscription tier is at least the specified tier
     * @param minimumTier - The minimum tier to check against
     * @returns Whether the current tier meets or exceeds the minimum
     */
    isTierAtLeast: (minimumTier: SubscriptionTier): boolean => {
      const tierOrder: SubscriptionTier[] = [
        "free",
        "pro",
        "business",
        "enterprise",
      ];
      const currentTierForComparison =
        effectiveTier === "trial" ? "business" : effectiveTier;
      return (
        tierOrder.indexOf(currentTierForComparison) >=
        tierOrder.indexOf(minimumTier)
      );
    },
  };
};
