import type { Workspace } from "@/types/workspace";
import { isMobileAuthFlow } from "@/lib/mobile-auth";
import { isPublicPath } from "@/public-portal-routes";
import { DEFAULT_WORKSPACE_PATH } from "@/shared/routing/workspace";
import { getSafeCallbackUrl } from "./callback-url";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

type InvitationRedirectCandidate = {
  token?: string;
};

const WORKSPACE_INDEPENDENT_PATHS = new Set([
  "/account",
  "/profile",
  "/onboarding/create",
  "/onboarding/join",
  "/oauth/authorize",
  "/github/callback",
]);

const canContinueWithoutWorkspace = (callbackUrl: string) => {
  if (isMobileAuthFlow(callbackUrl)) return true;

  const { pathname } = new URL(callbackUrl, "https://cloud.fortyone.app");
  return WORKSPACE_INDEPENDENT_PATHS.has(pathname) || isPublicPath(pathname);
};

export const getRedirectUrl = (
  workspaces: Workspace[],
  invitations: InvitationRedirectCandidate[] = [],
  lastUsedWorkspaceId?: string,
  callbackUrl?: string,
) => {
  const safeCallbackUrl = getSafeCallbackUrl(callbackUrl);
  if (
    safeCallbackUrl &&
    (workspaces.length > 0 || canContinueWithoutWorkspace(safeCallbackUrl))
  ) {
    return safeCallbackUrl;
  }

  if (workspaces.length === 0) {
    const invitation = invitations.find((item) => item.token);
    if (invitation?.token) {
      return `/onboarding/join?token=${encodeURIComponent(invitation.token)}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];

  return buildWorkspaceUrl(activeWorkspace.slug);
};

export const buildWorkspaceUrl = (
  slug: string,
  path = DEFAULT_WORKSPACE_PATH,
) => {
  if (isFortyOneApp) {
    return `https://${slug}.fortyone.app${path}`;
  }

  return `/${slug}${path}`;
};

export const withWorkspacePath = (path: string, slug?: string) => {
  if (!slug || isFortyOneApp) {
    return path;
  }

  if (path.startsWith(`/${slug}`)) {
    return path;
  }

  if (path.startsWith("/")) {
    return `/${slug}${path}`;
  }

  return `/${slug}/${path}`;
};
