"use client";

import type { ReactNode } from "react";
import type {
  SubscriptionTier,
  TIER_LIMITS,
} from "@/lib/hooks/subscription-features";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

/**
 * A component that conditionally renders content based on subscription features
 *
 * FeatureGuard can restrict content based on:
 * - Feature availability: Use the `feature` prop to check if a specific feature is enabled in the current plan
 * - Minimum tier: Use the `minimumTier` prop to require a specific subscription tier or higher
 * - Usage limits: Use the `feature` with `count` props to check if adding one more would exceed limits
 *  - If the feature is not available, the component will show the fallback content
 */
export const FeatureGuard = ({
  feature,
  minimumTier,
  count,
  children,
  fallback = null,
}: {
  feature?: keyof typeof TIER_LIMITS.free;
  minimumTier?: SubscriptionTier;
  count?: number;
  children: ReactNode;
  fallback?: ReactNode;
}) => {
  const { hasFeature, isTierAtLeast, withinLimit } = useSubscriptionFeatures();

  // Check if the feature is available based on the provided criteria
  const isFeatureAvailable =
    (feature && hasFeature(feature)) ||
    (minimumTier && isTierAtLeast(minimumTier)) ||
    (feature && count !== undefined && withinLimit(feature, count));

  // If no criteria provided, show children
  if (!feature && !minimumTier && count === undefined) {
    return <>{children}</>;
  }

  return isFeatureAvailable ? <>{children}</> : <>{fallback}</>;
};
