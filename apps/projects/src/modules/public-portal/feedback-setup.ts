import type { Workspace } from "@/types";
import { buildWorkspaceUrl } from "@/utils";
import { withCallbackUrl } from "@/utils/callback-url";

export const FEEDBACK_SETTINGS_PATH = "/settings/workspace/feedback";

export const getFeedbackOnboardingPath = () =>
  withCallbackUrl("/onboarding/create", FEEDBACK_SETTINGS_PATH);

export const getFeedbackSignupPath = () =>
  withCallbackUrl("/signup?source=portal", getFeedbackOnboardingPath());

export const getFeedbackSetupHref = (
  workspaces: Workspace[],
  lastUsedWorkspaceId?: string,
) => {
  const adminWorkspaces = workspaces.filter(
    (workspace) => workspace.userRole === "admin",
  );
  const activeWorkspace =
    adminWorkspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ??
    adminWorkspaces.at(0);

  return activeWorkspace
    ? buildWorkspaceUrl(activeWorkspace.slug, FEEDBACK_SETTINGS_PATH)
    : getFeedbackOnboardingPath();
};
