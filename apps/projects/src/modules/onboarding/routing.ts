import { buildWorkspaceUrl } from "@/utils";
import { getSafeCallbackUrl, withCallbackUrl } from "@/utils/callback-url";
import { DEFAULT_WORKSPACE_PATH } from "@/shared/routing/workspace";
import { isMobileAuthPath } from "@/lib/mobile-auth";

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
  isMobileAuthPath(callbackUrl)
    ? callbackUrl!
    : buildWorkspaceUrl(
        workspaceSlug,
        getOnboardingCallbackPath(callbackUrl) ?? DEFAULT_WORKSPACE_PATH,
      );
