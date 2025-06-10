import {
  useFeatureFlagEnabled,
  useFeatureFlagVariantKey,
} from "posthog-js/react";

export const availableFlags = ["analytics_page"] as const;

export const useFeatureFlag = (flag: (typeof availableFlags)[number]) => {
  return useFeatureFlagEnabled(flag);
};

export const useVariant = (flag: (typeof availableFlags)[number]) => {
  return useFeatureFlagVariantKey(flag);
};
