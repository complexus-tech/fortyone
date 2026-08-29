import { buildWorkspaceUrl } from "@/utils";
import { getSafeCallbackUrl, withCallbackUrl } from "@/utils/callback-url";

const DEFAULT_WORKSPACE_PATH = "/my-work";

export const getOnboardingCallbackPath = (callbackUrl?: string | null) => {
  const safeCallbackUrl = getSafeCallbackUrl(callbackUrl);

  return safeCallbackUrl?.startsWith("/") ? safeCallbackUrl : undefined;
};

export const withOnboardingCallbackUrl = (
  path: string,
  callbackUrl?: string | null,
) => withCallbackUrl(path, getOnboardingCallbackPath(callbackUrl));

export const getOnboardingWorkspaceUrl = (
  workspaceSlug: string,
  callbackUrl?: string | null,
) =>
  buildWorkspaceUrl(
    workspaceSlug,
    getOnboardingCallbackPath(callbackUrl) ?? DEFAULT_WORKSPACE_PATH,
  );
