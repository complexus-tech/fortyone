import type { OnboardingTourProgress, OnboardingTourStatus } from "./types";

type OnboardingTourScope = Pick<
  OnboardingTourProgress,
  "tourKey" | "tourVersion"
>;

const ONBOARDING_TOUR_STATUSES = new Set<OnboardingTourStatus>([
  "active",
  "completed",
  "skipped",
]);

const normalizeIdentifiers = (value: unknown) => {
  if (!Array.isArray(value)) {
    return [];
  }

  return Array.from(
    new Set(
      value.filter(
        (identifier): identifier is string => typeof identifier === "string",
      ),
    ),
  );
};

export const normalizeOnboardingTourProgress = (
  value: unknown,
  fallbackScope: OnboardingTourScope,
): OnboardingTourProgress => {
  if (!value || typeof value !== "object") {
    throw new Error("Invalid onboarding tour progress response");
  }

  const progress = value as Record<string, unknown>;
  const status = ONBOARDING_TOUR_STATUSES.has(
    progress.status as OnboardingTourStatus,
  )
    ? (progress.status as OnboardingTourStatus)
    : "active";

  return {
    completedActionIds: normalizeIdentifiers(progress.completedActionIds),
    completedStepIds: normalizeIdentifiers(progress.completedStepIds),
    status,
    tourKey:
      typeof progress.tourKey === "string"
        ? progress.tourKey
        : fallbackScope.tourKey,
    tourVersion:
      typeof progress.tourVersion === "string"
        ? progress.tourVersion
        : fallbackScope.tourVersion,
  };
};
