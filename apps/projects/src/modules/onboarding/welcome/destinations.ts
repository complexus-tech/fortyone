import type { Workspace } from "@/types/workspace";
import {
  getOnboardingWorkspaceUrl,
  withOnboardingCallbackUrl,
} from "@/modules/onboarding/routing";

const IMPORT_WORKSPACE_PATH = "/settings/workspace/imports?from=onboarding";

type WelcomeWorkspace = Pick<Workspace, "id" | "slug" | "userRole">;

type WelcomeDestinations = {
  redirectUrl: string;
  importUrl?: string;
};

export const getWelcomeDestinations = (
  workspaces: readonly WelcomeWorkspace[],
  lastUsedWorkspaceId?: string,
  callbackUrl?: string,
): WelcomeDestinations => {
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ??
    workspaces.at(0);

  if (!activeWorkspace) {
    return {
      redirectUrl: withOnboardingCallbackUrl("/onboarding/create", callbackUrl),
    };
  }

  return {
    redirectUrl: getOnboardingWorkspaceUrl(activeWorkspace.slug, callbackUrl),
    ...(activeWorkspace.userRole === "admin"
      ? {
          importUrl: getOnboardingWorkspaceUrl(
            activeWorkspace.slug,
            IMPORT_WORKSPACE_PATH,
          ),
        }
      : {}),
  };
};
