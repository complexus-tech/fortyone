import { DEFAULT_WORKSPACE_PATH } from "@/shared/routing/workspace";
import { buildWorkspaceUrl } from "@/utils/workspace-url";
import { getOnboardingCallbackPath } from "./routing";

export type OnboardingStart = "task" | "import" | "examples" | "empty";

export const ONBOARDING_START_QUERY = "onboarding";
export const ONBOARDING_TASK_START = "task";
export const ONBOARDING_CALENDAR_PATH = "/settings/account/calendar";

const START_PATHS: Record<OnboardingStart, string> = {
  task: `/my-work?${ONBOARDING_START_QUERY}=${ONBOARDING_TASK_START}`,
  import: "/settings/workspace/imports?from=onboarding",
  examples: DEFAULT_WORKSPACE_PATH,
  empty: DEFAULT_WORKSPACE_PATH,
};

export const getOnboardingStartUrl = (
  workspaceSlug: string,
  start: OnboardingStart,
  callbackUrl?: string,
): string =>
  buildWorkspaceUrl(
    workspaceSlug,
    getOnboardingCallbackPath(callbackUrl) ?? START_PATHS[start],
  );
