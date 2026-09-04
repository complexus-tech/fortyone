export type OnboardingTourStatus = "active" | "completed" | "skipped";

export type OnboardingTourProgress = {
  completedActionIds: string[];
  completedStepIds: string[];
  status: OnboardingTourStatus;
  tourKey: string;
  tourVersion: string;
};

export type UpdateOnboardingTourProgress = Pick<
  OnboardingTourProgress,
  "tourKey" | "tourVersion"
> & {
  completedActionIds?: string[];
  completedStepIds?: string[];
  status?: OnboardingTourStatus;
};
