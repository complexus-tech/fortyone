import type { Workspace } from "@/types/workspace";
import { isMobileAuthPath } from "@/lib/mobile-auth";
import {
  getOnboardingWorkspaceUrl,
  withOnboardingCallbackUrl,
} from "../routing";
import { getOnboardingStartUrl, ONBOARDING_CALENDAR_PATH } from "../start";

type WelcomeWorkspace = Pick<Workspace, "id" | "slug" | "userRole">;

type WelcomeDestinations = {
  redirectUrl: string;
  taskUrl?: string;
  importUrl?: string;
  calendarUrl?: string;
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

  // Mobile onboarding must finish the browser handoff, not enter optional web
  // setup pages with their own navigation, billing, or upgrade surfaces.
  if (isMobileAuthPath(callbackUrl)) return { redirectUrl: callbackUrl! };

  return {
    redirectUrl: getOnboardingWorkspaceUrl(activeWorkspace.slug, callbackUrl),
    calendarUrl: getOnboardingWorkspaceUrl(
      activeWorkspace.slug,
      ONBOARDING_CALENDAR_PATH,
    ),
    ...(activeWorkspace.userRole === "admin" ||
    activeWorkspace.userRole === "member"
      ? { taskUrl: getOnboardingStartUrl(activeWorkspace.slug, "task") }
      : {}),
    ...(activeWorkspace.userRole === "admin"
      ? {
          importUrl: getOnboardingStartUrl(activeWorkspace.slug, "import"),
        }
      : {}),
  };
};
